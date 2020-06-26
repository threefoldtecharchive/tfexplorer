package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
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

const timeLayout = "02-Jan-2006 15:04:05"

func printResult(r workloads.Workloader) {
	output := m{}
	fmt.Printf("ID: %d", r.GetID())
	fmt.Printf("Workload type: %s", r.GetWorkloadType().String())

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
	deleted bool
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
		if res.GetResult().State == workloads.ResultStateDeleted {
			continue
		}
		if res.GetCustomerTid() != job.user {
			continue
		}
		cOut <- res
	}
}

func getResult(id int) (res workloads.Workloader, err error) {
	return bcdb.Workloads.Get(schema.ID(id))
}
