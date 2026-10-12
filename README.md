# Mihomo RuleSets

Скомпилированные бинарные правила (`.mrs`) и текстовые списки (`.list`) для Mihomo / Clash Meta.

Некоторые правила были взяты из этих репозиториев:
- [MetaCubeX](https://github.com/MetaCubeX/meta-rules-dat)
- [Davoyan](https://github.com/Davoyan/mihomo-rule-sets)

Кастомные списки будут дополняться.

Пример использования:

```YAML
anchors:
  a1: &domain { type: http, format: mrs, behavior: domain, interval: 86400 }
  a2: &ipcidr { type: http, format: mrs, behavior: ipcidr, interval: 86400 }

rule-providers:
  telegram@ipcidr: { <<: *ipcidr, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/telegram@ipcidr.mrs }
  telegram@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/telegram@domain.mrs }
  valve@ipcidr: { <<: *ipcidr, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/valve@ipcidr.mrs }
  sonos@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/sonos@domain.mrs }
  spotify@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/spotify@domain.mrs }
  instagram@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/instagram@domain.mrs }
  mozilla@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/mozilla@domain.mrs }
  tiktok-additional@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/tiktok-additional@domain.mrs }
  keenetic@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/keenetic@domain.mrs }
  netcraze@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/netcraze@domain.mrs }
  deepl@domain: { <<: *domain, url: https://github.com/Deshaveu/mihomo-rulesets/releases/latest/download/deepl@domain.mrs }

rules:
  - RULE-SET,telegram@domain,PROXY
  - RULE-SET,telegram@ipcidr,PROXY
  - RULE-SET,valve@ipcidr,DIRECT
  - RULE-SET,sonos@domain,PROXY
  - RULE-SET,instagram@domain,PROXY
  - RULE-SET,spotify@domain,PROXY
  - RULE-SET,mozilla@domain,DIRECT
  - RULE-SET,tiktok-additional@domain,DIRECT
  - RULE-SET,keenetic@domain,DIRECT
  - RULE-SET,netcraze@domain,DIRECT
  - RULE-SET,deepl@domain,PROXY
  - MATCH,DIRECT
