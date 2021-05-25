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
	host     string
	database string
	user     string
	password string
}

const ProjectEntryDB = "projects"

func newSqlDbClient(host, database, user, password string) (sqlDbClient, error) {
	return sqlDbClient{
		host:     host,
		database: database,
		user:     user,
		password: password,
	}, nil
}

func (d sqlDbClient) createSession() (db.Session, error) {
	settings := postgresql.ConnectionURL{
		Host:     d.host,
		Database: d.database,
		User:     d.user,
		Password: d.password,
	}

	return postgresql.Open(settings)
}

func (d sqlDbClient) CreateProjectEntry(pe ProjectEntry) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Tx(func(sess db.Session) error {
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

	sess, err := d.createSession()
	if err != nil {
		return res, err
	}
	defer sess.Close()

	err = sess.Collection(ProjectEntryDB).Find("project", project).One(&res)
	return res, err
}

func (d sqlDbClient) DeleteProjectEntry(project string) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Collection(ProjectEntryDB).Find("project", project).Delete()
}
