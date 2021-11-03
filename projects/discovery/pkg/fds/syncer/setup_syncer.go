package syncer

import (
	"github.com/solo-io/gloo/pkg/utils/setuputils"
	discoveryRegistry "github.com/solo-io/gloo/projects/discovery/pkg/fds/discoveries/registry"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/enterprise/options/graphql/v1alpha1"
	"github.com/solo-io/gloo/projects/gloo/pkg/syncer/setup"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/gloo/projects/discovery/pkg/fds"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins/registry"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errutils"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

type Extensions struct {
	DiscoveryFactoryFuncs []func() fds.FunctionDiscoveryFactory
}

func NewSetupFunc() setuputils.SetupFunc {
	return setup.NewSetupFuncWithRunAndExtensions(RunFDS, nil)
}

func NewSetupFuncWithExtensions(extensions Extensions) setuputils.SetupFunc {
	runWithExtensions := func(opts bootstrap.Opts) error {
		return RunFDSWithExtensions(opts, extensions)
	}
	return setup.NewSetupFuncWithRunAndExtensions(runWithExtensions, nil)
}

func RunFDS(opts bootstrap.Opts) error {
	return RunFDSWithExtensions(opts, Extensions{})
}

func RunFDSWithExtensions(opts bootstrap.Opts, extensions Extensions) error {
	fdsMode := getFdsMode(opts.Settings)
	if fdsMode == v1.Settings_DiscoveryOptions_DISABLED {
		contextutils.LoggerFrom(opts.WatchOpts.Ctx).Info("function discovery disabled. to enable, modify "+
			"gloo.solo.io/Settings %v", opts.Settings.GetMetadata().Ref())
		return nil
	}

	watchOpts := opts.WatchOpts.WithDefaults()
	watchOpts.Ctx = contextutils.WithLogger(watchOpts.Ctx, "fds")

	upstreamClient, err := v1.NewUpstreamClient(watchOpts.Ctx, opts.Upstreams)
	if err != nil {
		return err
	}
	if err := upstreamClient.Register(); err != nil {
		return err
	}
	secretClient, err := v1.NewSecretClient(watchOpts.Ctx, opts.Secrets)
	if err != nil {
		return err
	}
	if err := secretClient.Register(); err != nil {
		return err
	}
	graphqlClient, err := v1alpha1.NewGraphQLSchemaClient(watchOpts.Ctx, opts.GraphQLSchemas)
	if err != nil {
		return err
	}
	if err := graphqlClient.Register(); err != nil {
		return err
	}

	var nsClient skkube.KubeNamespaceClient
	if opts.KubeClient != nil && opts.KubeCoreCache.NamespaceLister() != nil {
		nsClient = namespace.NewNamespaceClient(opts.KubeClient, opts.KubeCoreCache)
	} else {
		nsClient = &FakeKubeNamespaceWatcher{}
	}

	cache := v1.NewDiscoveryEmitter(upstreamClient, nsClient, secretClient)

	var resolvers fds.Resolvers
	for _, plug := range registry.Plugins(opts) {
		resolver, ok := plug.(fds.Resolver)
		if ok {
			resolvers = append(resolvers, resolver)
		}
	}

	// TODO: unhardcode
	functionalPlugins := GetFunctionDiscoveriesWithExtensions(opts, extensions)

	// TODO(yuval-k): max Concurrency here
	updater := fds.NewUpdater(watchOpts.Ctx, resolvers, graphqlClient, upstreamClient, 0, functionalPlugins)
	disc := fds.NewFunctionDiscovery(updater)

	sync := NewDiscoverySyncer(disc, fdsMode)
	eventLoop := v1.NewDiscoveryEventLoop(cache, sync)

	errs := make(chan error)

	eventLoopErrs, err := eventLoop.Run(opts.WatchNamespaces, watchOpts)
	if err != nil {
		return err
	}
	go errutils.AggregateErrs(watchOpts.Ctx, errs, eventLoopErrs, "event_loop.fds")

	logger := contextutils.LoggerFrom(watchOpts.Ctx)

	go func() {

		for {
			select {
			case err := <-errs:
				logger.Errorf("error: %v", err)
			case <-watchOpts.Ctx.Done():
				return
			}
		}
	}()
	return nil
}

func GetFunctionDiscoveriesWithExtensions(opts bootstrap.Opts, extensions Extensions) []fds.FunctionDiscoveryFactory {
	return GetFunctionDiscoveriesWithExtensionsAndRegistry(opts, discoveryRegistry.Plugins, extensions)
}

func GetFunctionDiscoveriesWithExtensionsAndRegistry(opts bootstrap.Opts, registryDiscFacts func(opts bootstrap.Opts) []fds.FunctionDiscoveryFactory, extensions Extensions) []fds.FunctionDiscoveryFactory {
	pluginfuncs := extensions.DiscoveryFactoryFuncs
	upgradedDiscoveryFactories := make(map[string]bool)
	discFactories := registryDiscFacts(opts)
	for _, discoveryFactoryExtension := range pluginfuncs {
		pe := discoveryFactoryExtension()
		if uPlug, ok := pe.(fds.Upgradable); ok && uPlug.IsUpgrade() {
			upgradedDiscoveryFactories[uPlug.FunctionDiscoveryFactoryName()] = true
		}
		discFactories = append(discFactories, pe)
	}
	return reconcileUpgradedDiscoveryFactories(discFactories, upgradedDiscoveryFactories)
}

func reconcileUpgradedDiscoveryFactories(factoryList []fds.FunctionDiscoveryFactory, upgradedFactories map[string]bool) []fds.FunctionDiscoveryFactory {
	var factoriesWithUpgrades []fds.FunctionDiscoveryFactory
	for _, fac := range factoryList {
		if uFactory, upgradable := fac.(fds.Upgradable); upgradable {
			_, hasRegisteredUpgrade := upgradedFactories[uFactory.FunctionDiscoveryFactoryName()]
			if (hasRegisteredUpgrade && uFactory.IsUpgrade()) || !hasRegisteredUpgrade {
				// if this is the upgraded version of the plugin
				// or this plugin does not have a registered upgrade
				// then add it to the final factory list
				factoriesWithUpgrades = append(factoriesWithUpgrades, fac)
			}
		} else {
			// if it's not an upgradable plugin, we don't need to worry about replacing it with an upgrade
			// so we can add it to the final factory list
			factoriesWithUpgrades = append(factoriesWithUpgrades, fac)
		}
	}
	return factoriesWithUpgrades
}

func getFdsMode(settings *v1.Settings) v1.Settings_DiscoveryOptions_FdsMode {
	if settings == nil || settings.GetDiscovery() == nil {
		return v1.Settings_DiscoveryOptions_WHITELIST
	}
	return settings.GetDiscovery().GetFdsMode()
}

// TODO: consider using regular solo-kit namespace client instead of KubeNamespace client
// to eliminate the need for this fake client for non kube environments
type FakeKubeNamespaceWatcher struct{}

func (f *FakeKubeNamespaceWatcher) Watch(opts clients.WatchOpts) (<-chan skkube.KubeNamespaceList, <-chan error, error) {
	return nil, nil, nil
}
func (f *FakeKubeNamespaceWatcher) BaseClient() clients.ResourceClient {
	return nil

}
func (f *FakeKubeNamespaceWatcher) Register() error {
	return nil
}
func (f *FakeKubeNamespaceWatcher) Read(name string, opts clients.ReadOpts) (*skkube.KubeNamespace, error) {
	return nil, nil
}
func (f *FakeKubeNamespaceWatcher) Write(resource *skkube.KubeNamespace, opts clients.WriteOpts) (*skkube.KubeNamespace, error) {
	return nil, nil
}
func (f *FakeKubeNamespaceWatcher) Delete(name string, opts clients.DeleteOpts) error {
	return nil
}
func (f *FakeKubeNamespaceWatcher) List(opts clients.ListOpts) (skkube.KubeNamespaceList, error) {
	return nil, nil
}
