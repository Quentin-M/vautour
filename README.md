# Vautour

By the original author of [Clair](https://github.com/coreos/clair/), Vautour is
a distributed & extensible web hunter. Crawling the internet, Vautour lists,
scrapes, processes (e.g. YARA) & persists documents asynchronously, looking
for content that may be of interest for organizations or security researchers.

### Supported Modules

Below are the modules currently supported by Vautour. Contributing new modules
is straight-forward, as it merely requires to implement the desired interface
as a new Go package in `src/modules` & importing it in `cmd/vautour/main.go` 
(or within your custom main file).

| Name           | Status | Notes
|----------------|--------|-----------------------------------|
| **Inputs**     |        |                                   |
| Pastebin       | ✅     | (Requires Pastebin PRO)           |
| Github / Gists | 🕒     | (Planned)                         |
| Stack Exchange | 🕒     | (Planned)                         |
| **Processors*  |        |                                   |
| YARA           | ✅     | ([Examples rules](config/rules/)) |
| **Outputs**    |        |                                   |
| ElasticSearch  | ✅     |                                   |
| **Queue**      |        |                                   |
| Redis          | ✅     |                                   |

### Getting started

- Read & acknowledge its [DISCLAIMER](DISCLAIMER), as well its [LICENSE](LICENSE)
- Run `docker-compose up`
- Wait a minute for the ELK stack to start, and for the first documents to be published
    - In the meantime, take a look at the default [config](config/vautour.yaml)
- Head to [Kibana](http://127.0.0.1:5601)
- Create an Index Pattern:
    - Name it "Vautour"
    - Choose "CreatedAt" as the time field,
    - Edit the "Content" field, set the format to "String" and the transform to "Base64 Decode"
- Profit.
    - Documents that matched the examples rules will have their `Score: >0`

### Roadmap

- Inputs: Github / Gists, Stack Exchange
- Outputs: MinScore support