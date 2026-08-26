package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"os"
	"regexp"
	"strings"
)

type mapping struct {
	StandardKey  string `json:"standard_key"`
	MediaAssetID string `json:"media_asset_id"`
}
type manifest struct {
	Version  int       `json:"version"`
	Mappings []mapping `json:"mappings"`
}

var keyRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var idRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func load(path string) (manifest, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return manifest{}, e
	}
	var m manifest
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if e = d.Decode(&m); e != nil {
		return m, e
	}
	if e = d.Decode(&struct{}{}); !errors.Is(e, io.EOF) {
		return m, errors.New("multiple JSON values")
	}
	return m, nil
}
func validate(m manifest) error {
	if m.Version != 1 {
		return errors.New("version must be 1")
	}
	seen := map[string]bool{}
	for _, x := range m.Mappings {
		if !keyRE.MatchString(x.StandardKey) || !idRE.MatchString(x.MediaAssetID) || seen[x.StandardKey] {
			return fmt.Errorf("invalid or duplicate mapping: %s", x.StandardKey)
		}
		seen[x.StandardKey] = true
	}
	return nil
}
func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	file := fs.String("file", "", "mapping manifest")
	valid := fs.Bool("validate-only", false, "validate only")
	dry := fs.Bool("dry-run", false, "database plan only")
	confirm := fs.Bool("confirm", false, "attach media")
	if e := fs.Parse(os.Args[1:]); e != nil {
		return e
	}
	modes := 0
	for _, v := range []bool{*valid, *dry, *confirm} {
		if v {
			modes++
		}
	}
	if *file == "" || modes != 1 {
		return errors.New("--file and exactly one mode are required")
	}
	m, e := load(*file)
	if e != nil {
		return e
	}
	if e = validate(m); e != nil {
		return e
	}
	fmt.Printf("VALID: %d demo mappings\n", len(m.Mappings))
	if *valid {
		return nil
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, dsn)
	if e != nil {
		return e
	}
	defer pool.Close()
	type item struct {
		mapping
		Current *string
	}
	items := []item{}
	attach, unchanged, conflicts := 0, 0, 0
	for _, x := range m.Mappings {
		var current *string
		var ok bool
		e = pool.QueryRow(ctx, `SELECT e.demo_media_id::text,(m.id IS NOT NULL AND m.status='ready' AND m.mime_type IN ('video/mp4','video/webm','image/gif','image/jpeg','image/png','image/webp') AND m.size_bytes<=5242880 AND (m.type='image' OR m.duration_seconds BETWEEN 1 AND 6)) FROM exercises e LEFT JOIN media_assets m ON m.id=$2::uuid WHERE e.standard_key=$1 AND e.owner_user_id IS NULL`, x.StandardKey, x.MediaAssetID).Scan(&current, &ok)
		if e != nil || !ok {
			conflicts++
			continue
		}
		if current != nil && *current == x.MediaAssetID {
			unchanged++
		} else {
			attach++
			items = append(items, item{x, current})
		}
	}
	fmt.Printf("PLAN: attach=%d unchanged=%d conflicts=%d\n", attach, unchanged, conflicts)
	if *dry {
		return nil
	}
	if conflicts > 0 {
		return errors.New("conflicts block attachment")
	}
	tx, e := pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	for _, x := range items {
		tag, err := tx.Exec(ctx, `UPDATE exercises SET demo_media_id=$2::uuid WHERE standard_key=$1 AND owner_user_id IS NULL`, x.StandardKey, x.MediaAssetID)
		if err != nil || tag.RowsAffected() != 1 {
			return errors.New("attachment failed")
		}
	}
	return tx.Commit(ctx)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
