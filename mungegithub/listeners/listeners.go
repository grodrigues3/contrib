package listeners

import (
	"k8s.io/contrib/mungegithub/github"
	"github.com/spf13/cobra"
	"fmt"
	"k8s.io/contrib/mungegithub/features"
	"k8s.io/contrib/mungegithub/listeners/listenerutil"
	"k8s.io/kubernetes/pkg/util/sets"
)
type Listener interface {
	AddFlags(cmd *cobra.Command, config *github.Config)
	Name() string
	RequiredFeatures() []string
	Initialize(*github.Config) error
	ListenForIssueEvent()
	ListenForCommentEvent()
	ListenForPullRequestEvent()
}

var (
	listenerMap map[string]Listener

)

func GetAllListeners() []Listeners {
	out := []Listeners{}
	for _, listener := range listenerMap {
		out = append(out, listener)
	}
	return out
}

// GetActiveListener returns a slice of all listeners which both registered and
// were requested by the user
func GetActiveListener() []Listener {
	return listeners
}

// RequestedFeatures returns a list of all feature which should be enabled
// for the running listeners
func RequestedFeatures() []string {
	out := sets.NewString()
	for _, m := range GetActiveListener() {
		f := m.RequiredFeatures()
		out.Insert(f...)
	}
	return out.List()
}

// RegisterListener will check if a requested listener exists and add it to
// the list.
func RegisterListener(requestedListener []string) error {
	for _, name := range requestedListener {
		listener, found := listenerMap[name]
		if !found {
			return fmt.Errorf("couldn't find a listener named: %s", name)
		}
		listeners = append(listeners, listener)
	}
	return nil
}

// InitializeListener will call listener.Initialize() for the requested listeners.
func InitializeListener(config *github.Config, features *features.Features) error {
	for _, listener := range listeners {
		if err := listener.Initialize(config, features); err != nil {
			return err
		}
		glog.Infof(listenerutil.PrettyString(listener))
		glog.Infof("Initialized listener: %s", listener.Name())
	}
	return nil
}

// EachLoop will be called before we start a poll loop and will run the
// EachLoop function for all active listeners
func EachLoop() error {
	for _, listener := range listeners {
		if err := listener.EachLoop(); err != nil {
			return err
		}
	}
	return nil
}

// RegisterListener should be called in `init()` by each listener to make itself
// available by name
func RegisterListener(listener Listener) error {
	if _, found := listenerMap[listener.Name()]; found {
		return fmt.Errorf("a listener with that name (%s) already exists", listener.Name())
	}
	listenerMap[listener.Name()] = listener
	glog.Infof("Registered %#v at %s", listener, listener.Name())
	return nil
}

// RegisterListener will call RegisterListener but will be fatal on error
func RegisterListener(listener Listener) {
	if err := RegisterListener(listener); err != nil {
		glog.Fatalf("Failed to register listener: %s", err)
	}
}

// ListenIssue will call each activated listener with the given object
func ListenIssue() error {
	for _, listener := range listeners {
		listener.Listen(obj)
	}
	return nil
}
