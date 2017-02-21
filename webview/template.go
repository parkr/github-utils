package webview

import (
	"fmt"
	"html/template"
	"time"
)

var tmplFuncs = template.FuncMap{"since": tmplTimeSince}

func tmplTimeSince(when time.Time) string {
	year, month, day, hour, min, sec := timeDiff(time.Now(), when)
	if year > 0 {
		return fmt.Sprintf(
			"%d years, %d months, %d days, %d hours, %d mins and %d seconds old",
			year, month, day, hour, min, sec)
	}

	if month > 0 {
		return fmt.Sprintf(
			"%d months, %d days, %d hours, %d mins and %d seconds old",
			month, day, hour, min, sec)
	}

	if day > 0 {
		return fmt.Sprintf("%d days, %d hours, %d mins and %d seconds old",
			day, hour, min, sec)
	}

	if hour > 0 {
		return fmt.Sprintf("%d hours, %d mins and %d seconds old", hour, min, sec)
	}

	if min > 0 {
		return fmt.Sprintf("%d mins and %d seconds old", min, sec)
	}

	return fmt.Sprintf("%d seconds old", sec)
}

func timeDiff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

var headerTmpl = template.Must(template.New("header").Parse(`
<html>
<head>
<title>{{.Title}}</title>
</head>
<body>
<pre>
<a href="/assigned" title="Assigned">Assigned</a> / <a href="/mentions" title="Mentions">Mentions</a> / <a href="/authored" title="Authored">Authored</a>
`))

var footerTmpl = template.Must(template.New("footer").Parse(`
</pre>
</body>
</html>
`))

var errorTmpl = template.Must(template.New("error").Parse(`


<p><strong>SHOOT! We hit a snag: {{.}}</strong></p>

`))

var issuesTableTmpl *template.Template

func init() {
	issuesTableTmpl = template.Must(template.New("funcs").Funcs(tmplFuncs).New("issueTable").Parse(`
{{.Title}} – {{len .Issues}} conversations.
<table border="0" cellspacing="2" cellpadding="2">
    <tr><th>Issue</th><th>Repo</th><th>Age</th><th></th></tr>
{{range .Issues}}<tr>
    <td><a href="{{.HTMLURL}}" title="{{.Title}}">{{.Title}}</a></td>
    <td><a href="{{.Repository.HTMLURL}}" title="{{.Repository.FullName}} on GitHub">{{.Repository.FullName}}</a></td>
    <td>{{since .CreatedAt}}</td>
</tr>{{end}}
</table>
`))
}
