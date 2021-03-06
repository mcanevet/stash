package check

import (
	"fmt"

	cs "github.com/appscode/stash/client/clientset/versioned/typed/stash/v1alpha1"
	"github.com/appscode/stash/pkg/cli"
	"github.com/appscode/stash/pkg/eventer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	CheckEventComponent = "stash-check"
)

type Options struct {
	Namespace   string
	ResticName  string
	HostName    string
	SmartPrefix string
}

type Controller struct {
	k8sClient   kubernetes.Interface
	stashClient cs.StashV1alpha1Interface
	opt         Options
}

func New(k8sClient kubernetes.Interface, stashClient cs.StashV1alpha1Interface, opt Options) *Controller {
	return &Controller{
		k8sClient:   k8sClient,
		stashClient: stashClient,
		opt:         opt,
	}
}

func (c *Controller) Run() (err error) {
	restic, err := c.stashClient.Restics(c.opt.Namespace).Get(c.opt.ResticName, metav1.GetOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			eventer.CreateEventWithLog(
				c.k8sClient,
				CheckEventComponent,
				restic.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCheck,
				fmt.Sprintf("Check failed for pod %s, reason: %s\n", c.opt.HostName, err),
			)
		} else {
			eventer.CreateEventWithLog(
				c.k8sClient,
				CheckEventComponent,
				restic.ObjectReference(),
				core.EventTypeNormal,
				eventer.EventReasonSuccessfulCheck,
				fmt.Sprintf("Check successful for pod: %s\n", c.opt.HostName),
			)
		}
	}()

	secret, err := c.k8sClient.CoreV1().Secrets(c.opt.Namespace).Get(restic.Spec.Backend.StorageSecretName, metav1.GetOptions{})
	if err != nil {
		return
	}

	cli := cli.New("/tmp", false, c.opt.HostName)
	if err = cli.SetupEnv(restic.Spec.Backend, secret, c.opt.SmartPrefix); err != nil {
		return
	}

	err = cli.Check()
	return
}
