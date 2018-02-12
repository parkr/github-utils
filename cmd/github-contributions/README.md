# github-contributions

Track your contributions to various GitHub repositories.

Example output (fabricated content):

    ### Pushed (1)
     * [x] [jekyll/jekyll#6748](https://github.com/jekyll/jekyll/pull/6748) WIP: Add basic benchmarking to CI


    ### Shipped (3)
     * [x] [jekyll/jekyll#6747](https://github.com/jekyll/jekyll/pull/6747) Minimize calls to `Addressable::URI.parse`
     * [x] [jekyll/jekyll#6746](https://github.com/jekyll/jekyll/pull/6746) WIP: Experiment with replacing Addressable for URL filters
     * [x] [jekyll/jekyll#6745](https://github.com/jekyll/jekyll/pull/6745) Add document on releasing a new version


    ### Tracked (3)
     * [x] [jekyll/benchmarking#13](https://github.com/jekyll/benchmarking/issues/13) Some way to specify which version of Jekyll to use
     * [x] [jekyll/benchmarking#9](https://github.com/jekyll/benchmarking/issues/9) Use different date for each post
     * [x] [jekyll/benchmarking#4](https://github.com/jekyll/benchmarking/issues/4) Use Minima for theme?


    ### Contributed (8)
     * [x] [jekyll/jekyll-seo-tag#272](https://github.com/jekyll/jekyll-seo-tag/issues/272) Allow section wide canonical_url replacement
     * [x] [jekyll/jekyll#6762](https://github.com/jekyll/jekyll/issues/6762) {%!h(MISSING)ighlight language %!}(MISSING) with empty line throws an exception
     * [x] [jekyll/jekyll#6761](https://github.com/jekyll/jekyll/issues/6761) Cache `site.url` and related configuration when using `--incremental`
     * [x] [jekyll/jekyll#6758](https://github.com/jekyll/jekyll/issues/6758) Problem with the SSL CA cert while testing the website locally using github-pages.
     * [x] [jekyll/jekyll-feed#210](https://github.com/jekyll/jekyll-feed/issues/210) how to set up pagination for feed?
     * [x] [jekyll/jekyll-sitemap#205](https://github.com/jekyll/jekyll-sitemap/issues/205) url parameter not used
     * [x] [jekyll/jekyll-feed#207](https://github.com/jekyll/jekyll-feed/issues/207) Does 'id' override actually work?
     * [x] [jekyll/jekyll#6743](https://github.com/jekyll/jekyll/issues/6743) Option to run converters before liquid


    ### Reviewed (17)
     * [x] [jekyll/jekyll#6766](https://github.com/jekyll/jekyll/pull/6766) Return empty content without attempting conversion
     * [x] [jekyll/jekyll#6764](https://github.com/jekyll/jekyll/pull/6764) Fix some common typos
     * [x] [jekyll/jekyll#6763](https://github.com/jekyll/jekyll/pull/6763) Cache and retrieve escaped path components
     * [x] [jekyll/jekyll-mentions#55](https://github.com/jekyll/jekyll-mentions/pull/55) Use default Rake tasks and scripts
     * [x] [jekyll/jekyll-commonmark#19](https://github.com/jekyll/jekyll-commonmark/pull/19) Version with class
     * [x] [jekyll/directory#43](https://github.com/jekyll/directory/pull/43) Update Rubocop
     * [x] [jekyll/jekyll#6751](https://github.com/jekyll/jekyll/pull/6751) Remove links to Gists
     * [x] [jekyll/benchmarking#12](https://github.com/jekyll/benchmarking/pull/12) Nicer if condition
     * [x] [jekyll/benchmarking#11](https://github.com/jekyll/benchmarking/pull/11) Remove duplicate do
     * [x] [jekyll/benchmarking#10](https://github.com/jekyll/benchmarking/pull/10) Remove unused layouts
     * [x] [jekyll/jekyll#6750](https://github.com/jekyll/jekyll/pull/6750) Added Premonition plugin to liste of plugins
     * [x] [jekyll/jekyll-feed#209](https://github.com/jekyll/jekyll-feed/pull/209) Escape image URL
     * [x] [jekyll/benchmarking#7](https://github.com/jekyll/benchmarking/pull/7) Add more filters to index
     * [x] [jekyll/benchmarking#6](https://github.com/jekyll/benchmarking/pull/6) Remove Disqus
     * [x] [jekyll/benchmarking#3](https://github.com/jekyll/benchmarking/pull/3) Use nicer C-style for loop
     * [x] [jekyll/jekyll#6744](https://github.com/jekyll/jekyll/pull/6744) Add 'jekyll-fontello' to plugins
     * [x] [jekyll/jekyll#6740](https://github.com/jekyll/jekyll/pull/6740) Access document permalink attribute efficiently

## Usage

```console
$ github-contributions -login=parkr
# Last week's issues & PR's in each category.
```

```console
$ github-contributions -login=parkr -owner=my-org
# Last week's issues & PR's in each category for a specific repository owner.
```

```console
$ github-contributions -login=parkr -since=2018-01-01
# All issues & PR's since January 1, 2018.
```

## Authentication

Authentication occurs via a `.netrc` file, like this:

```netrc
machine api.github.com
  login parkr
  password mypersonalaccesstoken
```

You can generate a Personal Access Token [in your
settings](https://github.com/settings/tokens).

## Installation

```console
$ go install github.com/parkr/github-contributions/...
```

This will install the binary into your `$GOPATH/bin` (usually `~/go/bin`).

## License

MIT.
