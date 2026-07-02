package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v71/github"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type result struct {
	Repo    string `json:"repo"`
	Content string `json:"content"`
}

type options struct {
	noRepoName      bool
	jsonOutput      bool
	includeArchived bool
	includeForks    bool
	regex           string
	concurrency     int
}

func buildRootCmd() *cobra.Command {
	var opts options

	rootCmd := &cobra.Command{
		Use:           "github-cat OWNER PATH",
		Short:         "Cat a file path across all repos owned by a github org or user",
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			path := args[1]
			return run(cmd.Context(), cmd.OutOrStdout(), owner, path, opts)
		},
	}

	rootCmd.Flags().BoolVarP(&opts.noRepoName, "no-repo-name", "n", false, "Do not prefix lines with repo name")
	rootCmd.Flags().BoolVarP(&opts.jsonOutput, "json", "", false, "JSON output")
	rootCmd.Flags().BoolVarP(&opts.includeArchived, "include-archived", "", false, "Include archived repos (off by default)")
	rootCmd.Flags().BoolVarP(&opts.includeForks, "include-forks", "", false, "Include forked repos (off by default)")
	rootCmd.Flags().StringVarP(&opts.regex, "regex", "r", "", "Filter lines in file by regex")
	rootCmd.Flags().IntVarP(&opts.concurrency, "concurrency", "j", 20, "Number of repos to fetch in parallel")

	return rootCmd
}

func run(ctx context.Context, out io.Writer, owner, path string, opts options) error {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return errors.New("GITHUB_TOKEN environment variable is not set")
	}

	var re *regexp.Regexp
	if opts.regex != "" {
		var err error
		re, err = regexp.Compile(opts.regex)
		if err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	}

	if opts.concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1, got %d", opts.concurrency)
	}

	client := github.NewClient(nil).WithAuthToken(token)

	repos, err := listRepos(ctx, client, owner, opts)
	if err != nil {
		return err
	}

	results, err := fetchContents(ctx, client, owner, path, repos, re, opts.concurrency)
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		return printJSON(out, results)
	}
	return printText(out, results, !opts.noRepoName)
}

// fetchContents retrieves the file at path from each repo in parallel, bounded
// by concurrency. Output order matches the input repos slice, and the first
// error cancels the remaining fetches.
func fetchContents(ctx context.Context, client *github.Client, owner, path string, repos []string, re *regexp.Regexp, concurrency int) ([]result, error) {
	found := make([]*result, len(repos))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i, repo := range repos {
		g.Go(func() error {
			content, ok, err := getContent(ctx, client, owner, repo, path)
			if err != nil {
				return err
			}
			if ok {
				found[i] = &result{
					Repo:    fmt.Sprintf("%s/%s", owner, repo),
					Content: filterRegex(content, re),
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	results := make([]result, 0, len(repos))
	for _, r := range found {
		if r != nil {
			results = append(results, *r)
		}
	}
	return results, nil
}

// listRepos returns the repo names for owner, which may be an organization or
// a user. It queries the organization endpoint first and falls back to the
// user endpoint when the owner is not an organization.
func listRepos(ctx context.Context, client *github.Client, owner string, opts options) ([]string, error) {
	names, err := listOrgRepos(ctx, client, owner, opts)
	if isNotFound(err) {
		return listUserRepos(ctx, client, owner, opts)
	}
	return names, err
}

func listOrgRepos(ctx context.Context, client *github.Client, org string, opts options) ([]string, error) {
	listOpts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var names []string
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, listOpts)
		if err != nil {
			return nil, err
		}

		names = appendFilteredRepos(names, repos, opts)

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return names, nil
}

func listUserRepos(ctx context.Context, client *github.Client, user string, opts options) ([]string, error) {
	listOpts := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var names []string
	for {
		repos, resp, err := client.Repositories.ListByUser(ctx, user, listOpts)
		if err != nil {
			return nil, err
		}

		names = appendFilteredRepos(names, repos, opts)

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return names, nil
}

func appendFilteredRepos(names []string, repos []*github.Repository, opts options) []string {
	for _, repo := range repos {
		if !opts.includeArchived && repo.GetArchived() {
			continue
		}
		if !opts.includeForks && repo.GetFork() {
			continue
		}
		names = append(names, repo.GetName())
	}
	return names
}

func getContent(ctx context.Context, client *github.Client, owner, repo, path string) (content string, found bool, err error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		if isNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}

	if fileContent == nil {
		return "", false, nil
	}

	decoded, err := fileContent.GetContent()
	if err != nil {
		return "", false, err
	}

	return decoded, true, nil
}

func isNotFound(err error) bool {
	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		return errResp.Response.StatusCode == http.StatusNotFound
	}
	return false
}

func filterRegex(content string, re *regexp.Regexp) string {
	if re == nil {
		return content
	}

	var matched []string
	for line := range strings.SplitSeq(content, "\n") {
		if re.MatchString(line) {
			matched = append(matched, line)
		}
	}

	return strings.Join(matched, "\n")
}

func printText(out io.Writer, results []result, repoName bool) error {
	for _, r := range results {
		for _, line := range splitLines(r.Content) {
			var err error
			if repoName {
				_, err = fmt.Fprintf(out, "%s:%s\n", r.Repo, line)
			} else {
				_, err = fmt.Fprintf(out, "%s\n", line)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// splitLines mirrors Ruby's String#lines: a single trailing newline yields no
// extra empty element, but interior blank lines are preserved.
func splitLines(content string) []string {
	if content == "" {
		return nil
	}
	trimmed := strings.TrimSuffix(content, "\n")
	return strings.Split(trimmed, "\n")
}

func printJSON(out io.Writer, results []result) error {
	if results == nil {
		results = []result{}
	}
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(data))
	return err
}

func main() {
	cmd := buildRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
