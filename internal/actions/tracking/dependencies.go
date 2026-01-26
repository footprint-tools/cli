package tracking

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/domain"
	"github.com/footprint-tools/footprint-cli/internal/git"
	repodomain "github.com/footprint-tools/footprint-cli/internal/repo"
	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/footprint-tools/footprint-cli/internal/ui"
)

type Deps struct {
	// git
	GitIsAvailable func() bool
	RepoRoot       func(string) (string, error)
	OriginURL      func(string) (string, error)
	ListRemotes    func(string) ([]string, error)
	GetRemoteURL   func(string, string) (string, error)
	HeadCommit     func() (string, error)
	CurrentBranch  func() (string, error)
	CommitMessage  func() (string, error)
	CommitAuthor   func() (string, error)

	// repo
	DeriveID func(string, string) (repodomain.RepoID, error)

	// store
	DBPath       func() string
	OpenDB       func(string) (*sql.DB, error)
	InitDB       func(*sql.DB) error
	InsertEvent  func(*sql.DB, store.RepoEvent) error
	ListEvents   func(*sql.DB, store.EventFilter) ([]store.RepoEvent, error)
	MarkOrphaned func(repoID repodomain.RepoID) (int64, error)

	// io
	Printf  func(string, ...any) (int, error)
	Println func(...any) (int, error)
	Pager   func(string)

	// misc
	Now    func() time.Time
	Getenv func(string) string

	// export sync
	GetExportRepo  func() string
	HasRemote      func(string) bool
	PullExportRepo func(string) error
	PushExportRepo func(string) error
}

func DefaultDeps() Deps {
	return Deps{
		GitIsAvailable: git.IsAvailable,
		RepoRoot:       git.RepoRoot,
		OriginURL:      git.OriginURL,
		ListRemotes:    git.ListRemotes,
		GetRemoteURL:   git.GetRemoteURL,
		HeadCommit:     git.HeadCommit,
		CurrentBranch:  git.CurrentBranch,
		CommitMessage:  git.CommitMessage,
		CommitAuthor:   git.CommitAuthor,

		DeriveID: repodomain.DeriveID,

		DBPath:       store.DBPath,
		OpenDB:       store.Open, //nolint:staticcheck // TODO: refactor to use store.New() with *Store interface
		InitDB:       store.Init,
		InsertEvent:  store.InsertEvent,
		ListEvents:   store.ListEvents,
		MarkOrphaned: markOrphanedWrapper,

		Printf:  fmt.Printf,
		Println: fmt.Println,
		Pager:   ui.Pager,

		Now:    time.Now,
		Getenv: os.Getenv,

		GetExportRepo:  getExportRepo,
		HasRemote:      hasRemote,
		PullExportRepo: pullExportRepo,
		PushExportRepo: pushExportRepo,
	}
}

// markOrphanedWrapper opens the database, marks events as orphaned, and closes.
func markOrphanedWrapper(repoID repodomain.RepoID) (int64, error) {
	s, err := store.New(store.DBPath())
	if err != nil {
		return 0, err
	}
	defer func() { _ = s.Close() }()

	// Convert repo.RepoID to domain.RepoID (both are string aliases)
	domainRepoID := domain.RepoID(string(repoID))
	return s.MarkOrphaned(domainRepoID)
}
