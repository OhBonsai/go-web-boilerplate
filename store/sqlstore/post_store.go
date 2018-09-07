package sqlstore

import (
	"github.com/mattermost/mattermost-server/einterfaces"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/mattermost/mattermost-server/model"
)

func NewSqlPostStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.PostStore {
	s := &SqlPostStore{
		SqlStore:          sqlStore,
		metrics:           metrics,
		lastPostTimeCache: utils.NewLru(LAST_POST_TIME_CACHE_SIZE),
		lastPostsCache:    utils.NewLru(LAST_POSTS_CACHE_SIZE),
		maxPostSizeCached: model.POST_MESSAGE_MAX_RUNES_V1,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Post{}, "Posts").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("RootId").SetMaxSize(26)
		table.ColMap("ParentId").SetMaxSize(26)
		table.ColMap("OriginalId").SetMaxSize(26)
		table.ColMap("Message").SetMaxSize(model.POST_MESSAGE_MAX_BYTES_V2)
		table.ColMap("Type").SetMaxSize(26)
		table.ColMap("Hashtags").SetMaxSize(1000)
		table.ColMap("Props").SetMaxSize(8000)
		table.ColMap("Filenames").SetMaxSize(model.POST_FILENAMES_MAX_RUNES)
		table.ColMap("FileIds").SetMaxSize(150)
	}

	return s
}
