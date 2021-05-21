package main

import (
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
)

type ProjectEntry struct {
	ProjectId  string `db:"project"`
	Repository string `db:"repository"`
}

type dbClient interface {
	CreateProjectEntry(pe ProjectEntry) error
	ReadProjectEntry(project string) (ProjectEntry, error)
	DeleteProjectEntry(project string) error
}

type sqlDbClient struct {
	sess db.Session
}

const ProjectEntryDB = "projects"

func newSqlDbClient(host, database, user, password string) (sqlDbClient, error) {
	settings := postgresql.ConnectionURL{
		Host:     host,
		Database: database,
		User:     user,
		Password: password,
	}

	sess, err := postgresql.Open(settings)
	if err != nil {
		return sqlDbClient{}, err
	}

	return sqlDbClient{
		sess: sess,
	}, nil
}

func (d sqlDbClient) CreateProjectEntry(pe ProjectEntry) error {
	return d.sess.Tx(func(sess db.Session) error {
		err := sess.Collection(ProjectEntryDB).Find("project", pe.ProjectId).Delete()
		if err != nil {
			return err
		}
		_, err = sess.Collection(ProjectEntryDB).Insert(pe)
		if err != nil {
			return err
		}

		return nil
	})
}

func (d sqlDbClient) ReadProjectEntry(project string) (ProjectEntry, error) {
	res := ProjectEntry{}
	err := d.sess.Collection(ProjectEntryDB).Find("project", project).One(&res)
	return res, err
}

func (d sqlDbClient) DeleteProjectEntry(project string) error {
	return d.sess.Collection(ProjectEntryDB).Find("project", project).Delete()
}
