package ixia

import (
	"context"
	"encoding/json"

	topopb "github.com/google/kne/proto/topo"
	"github.com/google/kne/topo/node"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IxiaSpec struct {
	Config string `json:"config,omitempty"`
}

type Ixia struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IxiaSpec `json:"spec,omitempty"`
}

func New(pb *topopb.Node) (node.Interface, error) {
	cfg := defaults(pb)
	proto.Merge(cfg, pb)
	node.FixServices(cfg)
	return &Node{
		pb: cfg,
	}, nil
}

type Node struct {
	pb *topopb.Node
}

func (n *Node) Proto() *topopb.Node {
	return n.pb
}

func (n *Node) CreateNodeResource(ctx context.Context, kClient kubernetes.Interface, ns string) error {
	log.Infof("Create IxiaTG node resource %s\n", n.pb.Name)
	jsonConfig, err := json.Marshal(n.pb.Config)
	if err != nil {
		log.Fatal(err)
	}
	newIxia := &Ixia{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "network.keysight.com/v1alpha1",
			Kind:       "IxiaTG",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.pb.Name,
			Namespace: ns,
		},
		Spec: IxiaSpec{
			Config: string(jsonConfig),
		},
	}
	body, err := json.Marshal(newIxia)
	if err != nil {
		log.Fatal(err)
	}

	err = kClient.CoreV1().RESTClient().
		Post().
		AbsPath("/apis/network.keysight.com/v1alpha1").
		Namespace(ns).
		Resource("Ixiatgs").
		Body(body).
		Do(ctx).
		Error()
	if err != nil {
		log.Error(err)
		return nil
	}
	log.Info("Success")
	return nil
}

func (n *Node) DeleteNodeResource(ctx context.Context, kClient kubernetes.Interface, ns string) error {
	log.Infof("Delete IxiaTG node resource %s\n", n.pb.Name)
	err := kClient.CoreV1().RESTClient().
		Delete().
		AbsPath("/apis/network.keysight.com/v1alpha1").
		Namespace(ns).
		Resource("Ixiatgs").
		Name(n.pb.Name).
		Do(ctx).
		Error()
	if err != nil {
		log.Error(err)
		return nil
	}
	log.Info("Success")
	return nil
}

func defaults(pb *topopb.Node) *topopb.Node {
	return &topopb.Node{}
}

func init() {
	node.Register(topopb.Node_IxiaTG, New)
}
