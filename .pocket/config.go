package main

import (
	"github.com/fredrikaverpil/pocket/pk"
	"github.com/fredrikaverpil/pocket/tasks/github"
	"github.com/fredrikaverpil/pocket/tasks/golang"
	"github.com/fredrikaverpil/pocket/tasks/markdown"
)

var Config = &pk.Config{
	Auto: pk.Parallel(
		pk.WithOptions(
			markdown.Tasks(),
			pk.WithSkipPath("jobs"),
		),
		pk.WithOptions(
			golang.Tasks(),
			pk.WithSkipPath("jobs"),
		),
		pk.WithOptions(
			github.Tasks(),
			pk.WithFlags(
				github.WorkflowFlags{
					Platforms:          []github.Platform{github.Ubuntu},
					GoReleaserWorkflow: new(true),
				},
			),
		),
	),
	Plan: &pk.PlanConfig{
		Shims: pk.DefaultShimConfig(),
	},
}
