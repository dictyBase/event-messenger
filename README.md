# event-messenger
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)  
![Continuous integration](https://github.com/dictyBase/event-messenger/workflows/Build/badge.svg?branch=develop)
[![codecov](https://codecov.io/gh/dictyBase/event-messenger/branch/develop/graph/badge.svg)](https://codecov.io/gh/dictyBase/event-messenger)
[![Maintainability](https://api.codeclimate.com/v1/badges/b760838bd7baa776bffd/maintainability)](https://codeclimate.com/github/dictyBase/event-messenger/maintainability)  
![Last commit](https://badgen.net/github/last-commit/dictyBase/event-messenger/develop)  
[![Funding](https://badgen.net/badge/Funding/Rex%20L%20Chisholm,dictyBase,DCR/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=10024726&icde=0)

dictyBase server to handle events as a subscriber through Nats messaging.

## Available commands

```bash
NAME:
   event-messenger - Handle events from nats messaging

USAGE:
   event-messenger [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     gh-issue    creates a github issue when a new stock order comes through
     send-email  sends an email when a new stock order comes through
     start-onto-server  starts the webhook server for loading ontologies
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-format value  format of the logging out, either of json or text. (default: "json")
   --log-level value   log level for the application (default: "error")
   --help, -h          show help
   --version, -v       print the version
```

## Misc badges

![Issues](https://badgen.net/github/issues/dictyBase/event-messenger)
![Open Issues](https://badgen.net/github/open-issues/dictyBase/event-messenger)
![Closed Issues](https://badgen.net/github/closed-issues/dictyBase/event-messenger)  
![Total PRS](https://badgen.net/github/prs/dictyBase/event-messenger)
![Open PRS](https://badgen.net/github/open-prs/dictyBase/event-messenger)
![Closed PRS](https://badgen.net/github/closed-prs/dictyBase/event-messenger)
![Merged PRS](https://badgen.net/github/merged-prs/dictyBase/event-messenger)  
![Commits](https://badgen.net/github/commits/dictyBase/event-messenger/develop)
![Branches](https://badgen.net/github/branches/dictyBase/event-messenger)
![Tags](https://badgen.net/github/tags/dictyBase/event-messenger/?color=cyan)  
![GitHub repo size](https://img.shields.io/github/repo-size/dictyBase/event-messenger?style=plastic)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/dictyBase/event-messenger?style=plastic)
[![Lines of Code](https://badgen.net/codeclimate/loc/dictyBase/event-messenger)](https://codeclimate.com/github/dictyBase/event-messenger/code)

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="http://cybersiddhu.github.com/"><img src="https://avatars.githubusercontent.com/u/48740?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Siddhartha Basu</b></sub></a><br /><a href="https://github.com/dictyBase/event-messenger/issues?q=author%3Acybersiddhu" title="Bug reports">🐛</a> <a href="https://github.com/dictyBase/event-messenger/commits?author=cybersiddhu" title="Code">💻</a> <a href="#content-cybersiddhu" title="Content">🖋</a> <a href="https://github.com/dictyBase/event-messenger/commits?author=cybersiddhu" title="Documentation">📖</a> <a href="#maintenance-cybersiddhu" title="Maintenance">🚧</a> <a href="#mentoring-cybersiddhu" title="Mentoring">🧑‍🏫</a></td>
    <td align="center"><a href="http://www.erichartline.net/"><img src="https://avatars.githubusercontent.com/u/13489381?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Eric Hartline</b></sub></a><br /><a href="https://github.com/dictyBase/event-messenger/issues?q=author%3Awildlifehexagon" title="Bug reports">🐛</a> <a href="https://github.com/dictyBase/event-messenger/commits?author=wildlifehexagon" title="Code">💻</a> <a href="#content-wildlifehexagon" title="Content">🖋</a> <a href="https://github.com/dictyBase/event-messenger/commits?author=wildlifehexagon" title="Documentation">📖</a> <a href="#maintenance-wildlifehexagon" title="Maintenance">🚧</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!