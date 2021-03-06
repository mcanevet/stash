package server

import (
	"fmt"
	"io"
	"net"

	"github.com/appscode/stash/pkg/controller"
	"github.com/appscode/stash/pkg/server"
	"github.com/spf13/pflag"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/stash.appscode.com"

type StashOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	ControllerOptions  *ControllerOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewStashOptions(out, errOut io.Writer) *StashOptions {
	o := &StashOptions{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, server.Codecs.LegacyCodec(admissionv1beta1.SchemeGroupVersion)),
		ControllerOptions:  NewControllerOptions(),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil

	return o
}

func (o StashOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	o.ControllerOptions.AddFlags(fs)
}

func (o StashOptions) Validate(args []string) error {
	return nil
}

func (o *StashOptions) Complete() error {
	return nil
}

func (o StashOptions) Config() (*server.StashConfig, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(server.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	controllerConfig := controller.NewControllerConfig(serverConfig.ClientConfig)
	if err := o.ControllerOptions.ApplyTo(controllerConfig); err != nil {
		return nil, err
	}

	config := &server.StashConfig{
		GenericConfig:    serverConfig,
		ControllerConfig: controllerConfig,
	}
	return config, nil
}

func (o StashOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}

	return s.Run(stopCh)
}
