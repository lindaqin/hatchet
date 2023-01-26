// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

var (
	SEVERITIES = []string{"F", "E", "W", "I", "D", "D2"}
	SEVERITY_M = map[string]string{"F": "FATAL", "E": "ERROR", "W": "WARN", "I": "INFO",
		"D": "DEBUG", "D2": "DEBUG2"}
)

// GetLogTableTemplate returns HTML
func GetLogTableTemplate(attr string) (*template.Template, error) {
	html := getContentHTML()
	if attr == "slowops" {
		html += getSlowOpsLogsTable()
	} else {
		html += getLegacyLogsTable()
	}
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"getComponentOptions": func(item string) template.HTML {
			arr := []string{}
			comps := []string{"ACCESS", "ASIO", "COMMAND", "CONNPOOL", "CONTROL", "ELECTION", "FTDC", "INDEX", "INITSYNC", "NETWORK",
				"QUERY", "RECOVERY", "REPL", "SHARDING", "STORAGE", "WRITE"}
			for _, v := range comps {
				selected := ""
				if v == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v, selected, v))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"getSeverityOptions": func(item string) template.HTML {
			arr := []string{}
			for _, v := range SEVERITIES {
				selected := ""
				if v == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v, selected, SEVERITY_M[v]))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"highlightLog": func(log string, params ...string) template.HTML {
			return template.HTML(highlightLog(log, params...))
		}}).Parse(html)
}

func highlightLog(log string, params ...string) string {
	re := regexp.MustCompile(`("?(planSummary|errMsg)"?:\s?"?\w+"?)`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`((\d+ms$))`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`(("?(keysExamined|keysInserted|docsExamined|nreturned|nMatched|nModified|ndeleted|ninserted|reslen)"?:)\d+)`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	for _, param := range params {
		re = regexp.MustCompile("(?i)(" + param + ")")
		log = re.ReplaceAllString(log, "<mark>$1</mark>")
	}
	return log
}

func getSlowOpsLogsTable() string {
	template := ` 
<p/>
<div align='center'>
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>date</th>
			<th>S</th>
			<th>component</th>
			<th>context</th>
			<th>message</th>
		</tr>
{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td>{{ $value.Timestamp }}</td>
			<td>{{ $value.Severity }}</td>
			<td>{{ $value.Component }}</td>
			<td>{{ $value.Context }}</td>
			<td>{{ highlightLog $value.Message }}</td>
		</tr>
{{end}}
	</table>
	<div align='center'><hr/><p/>@simagix</div>
</div>
`
	return template
}

func getLegacyLogsTable() string {
	template := `
<br/>
<div style="float: left;">
	<label><i class="fa fa-leaf"></i></label>
	<select id='component'>
		<option value=''>select a component</option>
		{{getComponentOptions .Component}}
	</select>
</div>

<div style="float: left; padding: 0px 0px 0px 20px;">
	<label><i class="fa fa-exclamation"></i></label>
	<select id='severity'>
		<option value=''>select a severity</option>
		{{getSeverityOptions .Severity}}
	</select>
</div>

<div style="float: left; padding: 0px 0px 0px 20px;">
	<label><i class="fa fa-search"></i></label>
	<input id='context' type='text' value='{{.Context}}' size='30'/>
	<button id="find" onClick="findLogs()" class="button" style="float: right;">Find</button>
</div>

<p/>
<div>
{{ if .Logs }}
	{{if .HasMore}}
		<button onClick="javascript:location.href='{{.URL}}'; return false;"
			class="btn" style="float: right;"><i class="fa fa-arrow-right"></i></button>
	{{end}}
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>date</th>
			<th>S</th>
			<th>component</th>
			<th>context</th>
			<th>message</th>
		</tr>
	{{$search := .Context}}
	{{$seq := .Seq}}
	{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n $seq }}</td>
			<td>{{ $value.Timestamp }}</td>
			<td>{{ $value.Severity }}</td>
			<td>{{ $value.Component }}</td>
			<td>{{ $value.Context }}</td>
			<td>{{ highlightLog $value.Message $search }}</td>
		</tr>
	{{end}}
	</table>
	{{if .HasMore}}
		<button onClick="javascript:location.href='{{.URL}}'; return false;"
			class="btn" style="float: right;"><i class="fa fa-arrow-right"></i></button>
	{{end}}
<div align='center'><hr/><p/>{{.Summary}}</div>
{{end}}
</div>
<script>
	var input = document.getElementById("context");
	input.addEventListener("keypress", function(event) {
		if (event.key === "Enter") {
			event.preventDefault();
			document.getElementById("find").click();
		}
	});

	function findLogs() {
		var sel = document.getElementById('component')
		var component = sel.options[sel.selectedIndex].value;
		sel = document.getElementById('severity')
		var severity = sel.options[sel.selectedIndex].value;
		var context = document.getElementById('context').value
		window.location.href = '/hatchets/{{.Hatchet}}/logs?component='+component+'&severity='+severity+'&context='+context;
	}
</script>
`
	return template
}
