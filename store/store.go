package store

import "github.com/mattermost/mattermost-server/model"

type Store interface {
	Close()
}

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

type PostStore interface {
	Save(post *model.Post) StoreChannel
	Update(newPost *model.Post, oldPost *model.Post) StoreChannel
	Get(id string) StoreChannel
	GetSingle(id string) StoreChannel
	Delete(postId string, time int64, deleteByID string) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	PermanentDeleteByChannel(channelId string) StoreChannel
	GetPosts(channelId string, offset int, limit int, allowFromCache bool) StoreChannel
	GetFlaggedPosts(userId string, offset int, limit int) StoreChannel
	GetFlaggedPostsForTeam(userId, teamId string, offset int, limit int) StoreChannel
	GetFlaggedPostsForChannel(userId, channelId string, offset int, limit int) StoreChannel
	GetPostsBefore(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsAfter(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsSince(channelId string, time int64, allowFromCache bool) StoreChannel
	GetEtag(channelId string, allowFromCache bool) StoreChannel
	Search(teamId string, userId string, params *model.SearchParams) StoreChannel
	AnalyticsUserCountsWithPostsByDay(teamId string) StoreChannel
	AnalyticsPostCountsByDay(teamId string) StoreChannel
	AnalyticsPostCount(teamId string, mustHaveFile bool, mustHaveHashtag bool) StoreChannel
	ClearCaches()
	InvalidateLastPostTimeCache(channelId string)
	GetPostsCreatedAt(channelId string, time int64) StoreChannel
	Overwrite(post *model.Post) StoreChannel
	GetPostsByIds(postIds []string) StoreChannel
	GetPostsBatchForIndexing(startTime int64, endTime int64, limit int) StoreChannel
	PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	GetOldest() StoreChannel
	GetMaxPostSize() StoreChannel
}