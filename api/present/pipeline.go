package present

import (
	"github.com/concourse/atc"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/dbng"
	"github.com/concourse/atc/web"
	"github.com/tedsuo/rata"
)

func Pipeline(savedPipeline dbng.Pipeline) atc.Pipeline {
	pathForRoute, err := web.Routes.CreatePathForRoute(web.Pipeline, rata.Params{
		"team_name": savedPipeline.TeamName(),
		"pipeline":  savedPipeline.Name(),
	})

	if err != nil {
		panic("failed to generate url: " + err.Error())
	}

	return atc.Pipeline{
		ID:       savedPipeline.ID(),
		Name:     savedPipeline.Name(),
		TeamName: savedPipeline.TeamName(),
		URL:      pathForRoute,
		Paused:   savedPipeline.Paused(),
		Public:   savedPipeline.Public(),
		Groups:   savedPipeline.Config().Groups,
	}
}
func DBPipeline(savedPipeline db.SavedPipeline) atc.Pipeline {
	pathForRoute, err := web.Routes.CreatePathForRoute(web.Pipeline, rata.Params{
		"team_name": savedPipeline.TeamName,
		"pipeline":  savedPipeline.Name,
	})

	if err != nil {
		panic("failed to generate url: " + err.Error())
	}

	return atc.Pipeline{
		ID:       savedPipeline.ID,
		Name:     savedPipeline.Name,
		TeamName: savedPipeline.TeamName,
		URL:      pathForRoute,
		Paused:   savedPipeline.Paused,
		Public:   savedPipeline.Public,
		Groups:   savedPipeline.Config.Groups,
	}
}
