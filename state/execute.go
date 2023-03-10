package state

import (
	"context"
	"fmt"
	"time"

	"github.com/redhat-appstudio-qe/performance-toolkit/config"
	"github.com/redhat-appstudio-qe/performance-toolkit/metrics"
)


func ExecuteExperiment(ctx context.Context, inject_chaos config.Inject, probe config.Probe) {
	c1 := make(chan bool)
	c2 := make(chan bool)
	iterations := ctx.Value("ChaosIteration").(int)
	tick := ctx.Value("ProbeIntervalSecs").(int)
	
	closeMetrics, metricsInstance := metrics.StartCollection(ctx)

	ctx = context.WithValue(ctx, "closeMetrics", closeMetrics)
	ctx = context.WithValue(ctx, "metricsInstance", metricsInstance)


	go func ()  {
		for {
			c1 <- true
			fmt.Println("inject chaos")
			inject_chaos(ctx)
			time.Sleep(1 * time.Second)
		}
	}()

	go func ()  {
		for x := range time.Tick(time.Duration(tick) * time.Second) {
			c2 <- true
			fmt.Println("probe")
			ctx = context.WithValue(ctx, "time", x)
			probe(ctx)
		}
	}()

	for i:=0; i<=iterations; i++ {
		println("retrigger- %d", i)
		<-c1
		<-c2
	}

	defer close(closeMetrics)
	metricsInstance.PrintResults()
}


func Chaos(ctx context.Context){
	explist := ctx.Value("ExperimentList").([]config.Expirement)
	for i := 0; i < len(explist); i++ {
        fmt.Println("Running expirement:", explist[i].Name)
		ctx = context.WithValue(ctx, "ProbeIntervalSecs", explist[i].ProbeIntervalSecs)
		ctx = context.WithValue(ctx, "ChaosIteration", explist[i].ChaosIteration)
		ctx = explist[i].Before(ctx)
		ExecuteExperiment(ctx, explist[i].Inject, explist[i].Probe)
		explist[i].After(ctx)
    }
}


