package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type Database struct {
	DataFile string
	Bucket   string
}

type Post struct {
	TagName       string
	Status        string
	ExecTime      float64
	ExecStartTime time.Time
	ExecEndTime   time.Time
	ExecCommand   *exec.Cmd
}

type JsonPost struct {
	TagName       string    `json:"TagName"`
	Status        string    `json:"Status"`
	ExecTime      float64   `json:"ExecTime"`
	ExecStartTime time.Time `json:"ExecStartTime"`
	ExecEndTime   time.Time `json:"ExecEndTime"`
	ExecCommand   *exec.Cmd `json:"ExecCommand"`
}

type JsonPosts []JsonPost

func LoadDatabase(config DatabaseConfig, bucket string) Database {
	d := Database{config.Path, bucket}

	db, err := bolt.Open(d.DataFile, 0600, nil)
	if err != nil {
		Logger("fatal", "baku.database", err.Error())
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(d.Bucket))
		if err != nil {
			return err
		}
		return nil
	})

	return d
}

func (d Database) Setup(groups []JobGroupConfig) {
	tagNames := []string{}

	for _, group := range groups {
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")

			preCheck := d.CheckStatus(tagName)

			if preCheck.TagName == "" {
				post := &Post{
					TagName: tagName,
					Status:  "initialized",
				}

				d.ChangeStatus(tagName, post)

				msg := "task initialize"
				Logger("info", tagName, msg)
			}

			tagNames = append(tagNames, tagName)
		}
	}

	db, err := bolt.Open(d.DataFile, 0600, nil)
	if err != nil {
		Logger("fatal", "baku.database", err.Error())
	}

	delTasks := []string{}
	runningTasks := []string{}

	db.View(func(tx *bolt.Tx) error {
		var initdata JsonPost

		bucket := tx.Bucket([]byte(d.Bucket))

		bucket.ForEach(func(k, v []byte) error {
			json.Unmarshal(v, &initdata)

			checkContains := Contains(tagNames, initdata.TagName)
			if checkContains == false {
				delTasks = append(delTasks, initdata.TagName)
			}

			if initdata.Status == "running" {
				runningTasks = append(runningTasks, initdata.TagName)
			}
			return nil
		})
		return nil
	})

	db.Close()

	for _, delTask := range delTasks {
		d.DeleteTask(delTask)
		Logger("info", delTask, "task delete")
	}

	for _, runningTask := range runningTasks {
		abort := d.CheckStatus(runningTask)
		post := &Post{
			TagName:       abort.TagName,
			Status:        "aborted(running)",
			ExecTime:      abort.ExecTime,
			ExecStartTime: abort.ExecStartTime,
			ExecEndTime:   abort.ExecEndTime,
			ExecCommand:   abort.ExecCommand,
		}

		d.ChangeStatus(runningTask, post)

		Logger("warn", runningTask, "task aborted")
	}
}

func (d Database) CheckStatus(tagName string) JsonPost {
	var jsonPost JsonPost

	db, err := bolt.Open(d.DataFile, 0600, nil)
	if err != nil {
		Logger("fatal", "baku.database", err.Error())
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(d.Bucket))
		view := bucket.Get([]byte(tagName))

		json.Unmarshal(view, &jsonPost)
		return nil
	})

	return jsonPost
}

func (d Database) ChangeStatus(tagName string, post *Post) {
	db, err := bolt.Open(d.DataFile, 0600, nil)
	if err != nil {
		Logger("fatal", "baku.database", err.Error())
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(d.Bucket))

		encoded, err := json.Marshal(post)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(tagName), encoded)
	})
}

func (d Database) DeleteTask(tagName string) {
	db, err := bolt.Open(d.DataFile, 0600, nil)
	if err != nil {
		Logger("fatal", "baku.database", err.Error())
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(d.Bucket))

		return bucket.Delete([]byte(tagName))
	})
}
