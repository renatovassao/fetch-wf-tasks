package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"github.com/renatovassao/workfront"
)

type Login struct {
	User string
	Pass string
}

type Mail struct {
	Server string
	Port   int
	From   *Login
	To     []string
}

type Config struct {
	Workfront *Login
	Mail      *Mail
}

func main() {

	// read config file
	file, err := os.Open("config.json")
	check(err)
	defer file.Close()

	// decode config file
	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	check(err)

	// check for config errors
	if config == nil || config.Workfront == nil || config.Mail == nil || config.Mail.From == nil {
		log.Fatal("Bad config file")
	}

	// login at workfront api
	userId, err := workfront.Login(config.Workfront.User, config.Workfront.Pass)
	check(err)

	// find workfront projects that logged in user is owner
	query := url.Values{"ownerID": {userId}}
	projects, err := workfront.SearchProjects(query)
	check(err)

	// e-mail message body
	var body = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
	</head>
	<body>`

	// workfront base url task view
	var baseURL = "https://agencia.my.workfront.com/task/view"

	// range over projects
	for _, p := range projects {
		// find tasks that are not complete in project <p>
		q := url.Values{"projectID": {p.ID}, "percentComplete": {"100"}, "percentComplete_Mod": {"lt"}}
		tasks, err := workfront.SearchTasks(q)
		check(err)

		// go to next project if no task is found
		if len(tasks) == 0 {
			continue
		}

		// append project title in message body
		body += fmt.Sprintf(`
		<div style="margin:0 0 0 72pt;">
			<font size="3" face="Calibri,sans-serif">
				<span style="font-size:12pt;">
					<font size="2">
						<span style="font-size:11pt;" lang="pt-BR">&nbsp;</span>
					</font>
				</span>
			</font>
		</div>
		<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
			<li style="margin:0 0 0 36pt;">
				<font size="3" face="Calibri,sans-serif">
					<span style="font-size:12pt;">
						<font size="2">
							<span style="font-size:11pt;" lang="pt-BR">
								<b>%s</b>
							</span>
						</font>
					</span>
				</font>
			</li>
		</ul>
		<div style="margin:0;">
			<font size="3" face="Calibri,sans-serif">
				<span style="font-size:12pt;">&nbsp;</span>
			</font>
		</div>
		`, p.Name)

		// range over tasks
		for _, t := range tasks {
			// split planned completion date
			split := strings.Split(t.PlannedCompletionDate, "T")

			// workfront task link
			link := baseURL + "?ID=" + t.ID

			// append task info in message body
			body += fmt.Sprintf(`
		<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
			<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
				<li style="margin:0 0 0 36pt;">
					<font size="3" face="Calibri,sans-serif">
						<span style="font-size:12pt;">
							<font size="2">
								<span style="font-size:11pt;" lang="pt-BR">&nbsp;</span>
							</font>
							<font size="2">
								<span style="font-size:11pt;" lang="pt-BR">
									<b>%s</b>
								</span>
							</font>
							<font size="2">
								<span style="font-size:11pt;" lang="pt-BR">:: %s</span>
							</font>
						</span>
					</font>
				</li>
			</ul>
		</ul>
		<div style="margin:0 0 0 144pt;">
			<font size="3" face="Calibri,sans-serif">
				<span style="font-size:12pt;">
					<font size="2">
						<span style="font-size:11pt;" lang="pt-BR">&nbsp;</span>
					</font>
				</span>
			</font>
		</div>
		<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
			<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
				<ul style="margin-top:0;margin-bottom:0;list-style-type:disc;">
					<li style="margin:0 0 0 36pt;">
						<font size="3" face="Calibri,sans-serif">
							<span style="font-size:12pt;">
								<a href="%s" target="_blank" rel="noopener noreferrer" data-auth="NotApplicable">
									<font size="2">
										<span style="font-size:11pt;" lang="pt-BR">%s</span>
									</font>
								</a>
							</span>
						</font>
					</li>
				</ul>
			</ul>
		</ul>
		<div style="margin:0;">
			<font size="3" face="Calibri,sans-serif">
				<span style="font-size:12pt;">
					<font size="2">
						<span style="font-size:11pt;">&nbsp;</span>
					</font>
				</span>
			</font>
		</div>
			`, split[0], t.Name, link, link)
		}
	}

	// finish up message body
	body += `
	</body>
</html>`

	var to string
	for _, v := range config.Mail.To {
		to += fmt.Sprintf("%s; ", v)
	}

	msg := "From: " + config.Mail.From.User + "\n" +
		"To: " + to + "\n" +
		"Subject: Pending Workfront Tasks\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
		body

	err = smtp.SendMail(fmt.Sprintf("%s:%d", config.Mail.Server, config.Mail.Port),
		smtp.PlainAuth("", config.Mail.From.User, config.Mail.From.Pass, config.Mail.Server),
		config.Mail.From.User, config.Mail.To, []byte(msg))

	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
