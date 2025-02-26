package public

import (
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"time"
)

var (
	LevelDB *DatabaseStruct
)

func init() {
	// InitQueue = append(InitQueue, InitProcess{
	// 	Order:    1,
	// 	InitFunc: loadLevelDB,
	// 	QuitFunc: closeLevelDB,
	// })
}

type DatabaseStruct struct {
	db_obj *leveldb.DB
}

func loadLevelDB() {
	db_obj, err := leveldb.OpenFile("./datas", nil)

	if err != nil {
		Log.Critical(err.Error())
		time.Sleep(50 * time.Millisecond)
		os.RemoveAll("./datas")
		os.Exit(127)
	}

	LevelDB = &DatabaseStruct{
		db_obj: db_obj,
	}
}

func closeLevelDB() {
	LevelDB.Close()
}

func (l *DatabaseStruct) Get(key string) (value string) {
	bytes_key := []byte(key)
	bytes_value, err := l.db_obj.Get(bytes_key, nil)

	if err != nil {
		return ""
	}

	value = string(bytes_value)
	return
}

func (l *DatabaseStruct) Put(key, value string) {
	bytes_key := []byte(key)
	bytes_value := []byte(value)
	_ = l.db_obj.Put(bytes_key, bytes_value, nil)
}

func (l *DatabaseStruct) Delete(key string) {
	bytes_key := []byte(key)
	_ = l.db_obj.Delete(bytes_key, nil)
}

func (l *DatabaseStruct) GetBytes(key string) (value []byte) {
	bytes_key := []byte(key)
	value, err := l.db_obj.Get(bytes_key, nil)

	if err != nil {
		return nil
	}

	return
}

func (l *DatabaseStruct) PutBytes(key string, value []byte) {
	bytes_key := []byte(key)
	_ = l.db_obj.Put(bytes_key, value, nil)
}

func (l *DatabaseStruct) Close() (err error) {
	err = l.db_obj.Close()
	return
}
