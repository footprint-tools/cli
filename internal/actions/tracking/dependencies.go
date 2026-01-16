package tracking

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Skryensya/footprint/internal/git"
	repodomain "github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/telemetry"
	"github.com/Skryensya/footprint/internal/ui"
)

type Deps struct {
	// git
	GitIsAvailable func() bool
	RepoRoot       func(string) (string, error)
	OriginURL      func(string) (string, error)
	HeadCommit     func() (string, error)
	CurrentBranch  func() (string, error)
	CommitMessage  func() (string, error)

	// repo
	DeriveID    func(string, string) (repodomain.RepoID, error)
	Track       func(repodomain.RepoID) (bool, error)
	Untrack     func(repodomain.RepoID) (bool, error)
	IsTracked   func(repodomain.RepoID) (bool, error)
	ListTracked func() ([]repodomain.RepoID, error)

	// telemetry
	DBPath      func() string
	OpenDB      func(string) (*sql.DB, error)
	InitDB      func(*sql.DB) error
	InsertEvent func(*sql.DB, telemetry.RepoEvent) error
	ListEvents  func(*sql.DB, telemetry.EventFilter) ([]telemetry.RepoEvent, error)

	// io
	Printf  func(string, ...any) (int, error)
	Println func(...any) (int, error)
	Pager   func(string)

	// misc
	Now    func() time.Time
	Getenv func(string) string
}

func DefaultDeps() Deps {
	return Deps{
		GitIsAvailable: git.IsAvailable,
		RepoRoot:       git.RepoRoot,
		OriginURL:      git.OriginURL,
		HeadCommit:     git.HeadCommit,
		CurrentBranch:  git.CurrentBranch,
		CommitMessage:  git.CommitMessage,

		DeriveID:    repodomain.DeriveID,
		Track:       repodomain.Track,
		Untrack:     repodomain.Untrack,
		IsTracked:   repodomain.IsTracked,
		ListTracked: repodomain.ListTracked,

		DBPath:      telemetry.DBPath,
		OpenDB:      telemetry.Open,
		InitDB:      telemetry.Init,
		InsertEvent: telemetry.InsertEvent,
		ListEvents:  telemetry.ListEvents,

		Printf:  fmt.Printf,
		Println: fmt.Println,
		Pager:   ui.Pager,

		Now:    time.Now,
		Getenv: os.Getenv,
	}
}
