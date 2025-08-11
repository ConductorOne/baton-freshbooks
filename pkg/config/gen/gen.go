package main

import (
	cfg "github.com/conductorone/baton-freshbooks/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("freshbooks", cfg.Config)
}
