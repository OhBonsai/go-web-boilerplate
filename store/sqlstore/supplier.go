package sqlstore

import (
	"go-web-boilerplate/store"
	"go-web-boilerplate/model"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/einterfaces"
	"fmt"
	"time"
	"os"
	"go-web-boilerplate/mlog"
)

const (
	INDEX_TYPE_FULL_TEXT	 	= "full_text"
	INDEX_TYPE_DEFAULT   	= "default"
	MAX_DB_CONN_LIFETIME 	= 60
	DB_PING_ATTEMPTS     		= 18
	DB_PING_TIMEOUT_SECS 	= 10
)

const (
	EXIT_CREATE_TABLE                		= 100
	EXIT_DB_OPEN                     		= 101
	EXIT_PING                        		= 102
	EXIT_NO_DRIVER                   		= 103
	EXIT_TABLE_EXISTS                		= 104
	EXIT_TABLE_EXISTS_MYSQL          	= 105
	EXIT_COLUMN_EXISTS               	= 106
	EXIT_DOES_COLUMN_EXISTS_POSTGRES 	= 107
	EXIT_DOES_COLUMN_EXISTS_MYSQL    	= 108
	EXIT_DOES_COLUMN_EXISTS_MISSING  	= 109
	EXIT_CREATE_COLUMN_POSTGRES      	= 110
	EXIT_CREATE_COLUMN_MYSQL         	= 111
	EXIT_CREATE_COLUMN_MISSING       	= 112
	EXIT_REMOVE_COLUMN               	= 113
	EXIT_RENAME_COLUMN               	= 114
	EXIT_MAX_COLUMN                  		= 115
	EXIT_ALTER_COLUMN                		= 116
	EXIT_CREATE_INDEX_POSTGRES       	= 117
	EXIT_CREATE_INDEX_MYSQL          	= 118
	EXIT_CREATE_INDEX_FULL_MYSQL     	= 119
	EXIT_CREATE_INDEX_MISSING        	= 120
	EXIT_REMOVE_INDEX_POSTGRES       	= 121
	EXIT_REMOVE_INDEX_MYSQL          	= 122
	EXIT_REMOVE_INDEX_MISSING        	= 123
	EXIT_REMOVE_TABLE                		= 134
	EXIT_CREATE_INDEX_SQLITE         	= 135
	EXIT_REMOVE_INDEX_SQLITE         	= 136
	EXIT_TABLE_EXISTS_SQLITE         	= 137
	EXIT_DOES_COLUMN_EXISTS_SQLITE   	= 138
)
type SqlSupplier struct {
	// rrCounter and srCounter should be kept first.
	// See https://github.com/mattermost/mattermost-server/pull/7281
	rrCounter      int64
	srCounter      int64
	next           store.LayeredStoreSupplier
	master         *gorp.DbMap
	replicas       []*gorp.DbMap
	searchReplicas []*gorp.DbMap
	oldStores      SqlSupplierOldStores
	settings       *model.SqlSettings
}

type SqlSupplierOldStores struct {
	post                 store.PostStore
}

func NewSqlSupplier(settings model.SqlSettings, metrics einterfaces.MetricsInterface) *SqlSupplier {
	supplier := &SqlSupplier{
		rrCounter: 0,
		srCounter: 0,
		settings:  &settings,
	}

	supplier.initConnection()

	supplier.oldStores.post = NewSqlPostStore(supplier, metrics)

	//initSqlSupplierReactions(supplier)
	//initSqlSupplierRoles(supplier)
	//initSqlSupplierSchemes(supplier)

	err := supplier.GetMaster().CreateTablesIfNotExists()
	if err != nil {
		mlog.Critical(fmt.Sprintf("Error creating database tables: %v", err))
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_TABLE)
	}

	UpgradeDatabase(supplier)

	supplier.oldStores.post.(*SqlPostStore).CreateIndexesIfNotExists()


	return supplier
}

func setupConnection(con_type string, dataSource string, settings *model.SqlSettings) *gorp.DbMap {
	db, err := dbsql.Open(*settings.DriverName, dataSource)
	if err != nil {
		mlog.Critical(fmt.Sprintf("Failed to open SQL connection to err:%v", err.Error()))
		time.Sleep(time.Second)
		os.Exit(EXIT_DB_OPEN)
	}

	for i := 0; i < DB_PING_ATTEMPTS; i++ {
		mlog.Info(fmt.Sprintf("Pinging SQL %v database", con_type))
		ctx, cancel := context.WithTimeout(context.Background(), DB_PING_TIMEOUT_SECS*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			break
		} else {
			if i == DB_PING_ATTEMPTS-1 {
				mlog.Critical(fmt.Sprintf("Failed to ping DB, server will exit err=%v", err))
				time.Sleep(time.Second)
				os.Exit(EXIT_PING)
			} else {
				mlog.Error(fmt.Sprintf("Failed to ping DB retrying in %v seconds err=%v", DB_PING_TIMEOUT_SECS, err))
				time.Sleep(DB_PING_TIMEOUT_SECS * time.Second)
			}
		}
	}

	db.SetMaxIdleConns(*settings.MaxIdleConns)
	db.SetMaxOpenConns(*settings.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(MAX_DB_CONN_LIFETIME) * time.Minute)

	var dbmap *gorp.DbMap

	connectionTimeout := time.Duration(*settings.QueryTimeout) * time.Second

	if *settings.DriverName == model.DATABASE_DRIVER_SQLITE {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.SqliteDialect{}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_MYSQL {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_POSTGRES {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.PostgresDialect{}, QueryTimeout: connectionTimeout}
	} else {
		mlog.Critical("Failed to create dialect specific driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_NO_DRIVER)
	}

	if settings.Trace {
		dbmap.TraceOn("", sqltrace.New(os.Stdout, "sql-trace:", sqltrace.Lmicroseconds))
	}

	return dbmap
}


func (s *SqlSupplier) initConnection() {
	s.master = setupConnection("master", *s.settings.DataSource, s.settings)

	if len(s.settings.DataSourceReplicas) > 0 {
		s.replicas = make([]*gorp.DbMap, len(s.settings.DataSourceReplicas))
		for i, replica := range s.settings.DataSourceReplicas {
			s.replicas[i] = setupConnection(fmt.Sprintf("replica-%v", i), replica, s.settings)
		}
	}

	if len(s.settings.DataSourceSearchReplicas) > 0 {
		s.searchReplicas = make([]*gorp.DbMap, len(s.settings.DataSourceSearchReplicas))
		for i, replica := range s.settings.DataSourceSearchReplicas {
			s.searchReplicas[i] = setupConnection(fmt.Sprintf("search-replica-%v", i), replica, s.settings)
		}
	}
}