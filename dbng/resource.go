package dbng

import (
	"database/sql"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/concourse/atc"
)

//go:generate counterfeiter . Resource

type Resource interface {
	ID() int
	Name() string
	PipelineName() string
	Type() string
	Source() atc.Source
	CheckEvery() string
	Tags() atc.Tags
	CheckError() error
	Paused() bool
}

var resourcesQuery = psql.Select("r.id, r.name, r.config, r.check_error, r.paused, r.pipeline_id, p.name").
	From("resources r").
	Join("pipelines p ON p.id = r.pipeline_id").
	Where(sq.Eq{"r.active": true})

type resource struct {
	id           int
	name         string
	pipelineID   int
	pipelineName string
	type_        string
	source       atc.Source
	checkEvery   string
	tags         atc.Tags
	checkError   error
	paused       bool

	conn Conn
}

func (r *resource) ID() int              { return r.id }
func (r *resource) Name() string         { return r.name }
func (r *resource) PipelineID() int      { return r.pipelineID }
func (r *resource) PipelineName() string { return r.pipelineName }
func (r *resource) Type() string         { return r.type_ }
func (r *resource) Source() atc.Source   { return r.source }
func (r *resource) CheckEvery() string   { return r.checkEvery }
func (r *resource) Tags() atc.Tags       { return r.tags }
func (r *resource) CheckError() error    { return r.checkError }
func (r *resource) Paused() bool         { return r.paused }

func scanResource(r *resource, row scannable) error {
	var (
		configBlob []byte
		checkErr   sql.NullString
	)

	err := row.Scan(&r.id, &r.name, &configBlob, &checkErr, &r.paused, &r.pipelineID, &r.pipelineName)
	if err != nil {
		return err
	}

	var config atc.ResourceConfig
	err = json.Unmarshal(configBlob, &config)
	if err != nil {
		return err
	}

	r.type_ = config.Type
	r.source = config.Source
	r.checkEvery = config.CheckEvery
	r.tags = config.Tags

	if checkErr.Valid {
		r.checkError = errors.New(checkErr.String)
	}

	return nil
}
