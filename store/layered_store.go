package store

import (
	"context"
	"github.com/mattermost/mattermost-server/einterfaces"
)


type LayeredStore struct {
	TmpContext      context.Context
}


func NewLayeredStore(db LayeredStoreDatabaseLayer, metrics einterfaces.MetricsInterface, cluster einterfaces.ClusterInterface) Store {
	store := &LayeredStore{
		TmpContext:      context.TODO(),
	}

	return store
}