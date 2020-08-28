package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/workloads"
	"gopkg.in/yaml.v2"

	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/urfave/cli"
)

func cmdsLive(c *cli.Context) error {
	var (
		userID  = int64(mainui.ThreebotID)
		start   = c.Int("start")
		end     = c.Int("end")
		expired = c.Bool("expired")
		deleted = c.Bool("deleted")
	)

	s := scraper{
		poolSize: 10,
		start:    start,
		end:      end,
		expired:  expired,
		deleted:  deleted,
	}
	cResults := s.Scrap(userID)
	for result := range cResults {
		printResult(result)
	}
	return nil
}

type m map[string]interface{}

func printResult(r workloads.Workloader) {
	output := m{}
	c := r.GetContract()
	fmt.Printf("ID: %d", c.ID)
	fmt.Printf("Workload type: %s", c.WorkloadType.String())

	if err := yaml.NewEncoder(os.Stdout).Encode(output); err != nil {
		log.Error().Err(err).Msg("failed to print result")
	}

	fmt.Println("-------------------------")
}

type scraper struct {
	poolSize int
	start    int
	end      int
	expired  bool
	deleted  bool
	wg       sync.WaitGroup
}
type job struct {
	id      int
	user    int64
	expired bool
}

func (s *scraper) Scrap(user int64) chan workloads.Workloader {

	var (
		cIn  = make(chan job)
		cOut = make(chan workloads.Workloader)
	)

	s.wg.Add(s.poolSize)
	for i := 0; i < s.poolSize; i++ {
		go worker(&s.wg, cIn, cOut)
	}

	go func() {
		defer func() {
			close(cIn)
		}()
		for i := s.start; i < s.end; i++ {
			cIn <- job{
				id:      i,
				user:    user,
				expired: s.expired,
			}
		}
	}()

	go func() {
		s.wg.Wait()
		close(cOut)
	}()

	return cOut
}

func worker(wg *sync.WaitGroup, cIn <-chan job, cOut chan<- workloads.Workloader) {
	defer func() {
		wg.Done()
	}()

	for job := range cIn {
		res, err := getResult(job.id)
		if err != nil {
			continue
		}
		s := res.GetState()
		c := res.GetContract()
		if s.Result.State == workloads.ResultStateDeleted {
			continue
		}
		if c.CustomerTid != job.user {
			continue
		}
		cOut <- res
	}
}

func getResult(id int) (res workloads.Workloader, err error) {
	return bcdb.Workloads.Get(schema.ID(id))
}
