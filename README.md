# keylightctl

A command line tool for controlling Elgato [Key
Lights](https://www.elgato.com/en/gaming/key-light) and [Key Light
Airs](https://www.elgato.com/en/gaming/key-light-air).

A library for interacting with the lights yourself is available at
[endocrimes/keylight-go](https://github.com/endocrimes/keylight-go).

## Example

```bash
[keylightctl(master)] $ ./bin/keylightctl describe --all
+---+--------------------------+-------------+------------+-------------+
| # | NAME                     | POWER STATE | BRIGHTNESS | TEMPERATURE |
+---+--------------------------+-------------+------------+-------------+
| 0 | Elgato\ Key\ Light\ 861A | on          |         50 |         295 |
+---+--------------------------+-------------+------------+-------------+
[keylightctl(master)] $ ./bin/keylightctl switch --light 861A off               
[keylightctl(master)] $ ./bin/keylightctl describe --all         
+---+--------------------------+-------------+------------+-------------+
| # | NAME                     | POWER STATE | BRIGHTNESS | TEMPERATURE |
+---+--------------------------+-------------+------------+-------------+
| 0 | Elgato\ Key\ Light\ 861A | off         |         50 |         295 |
+---+--------------------------+-------------+------------+-------------+
[keylightctl(master)] $ ./bin/keylightctl switch --light 861A --brightness 25 on
[keylightctl(master)] $ ./bin/keylightctl describe --all                        
==> Found no matching lights during discovery
[keylightctl(master)] $ ./bin/keylightctl describe --all
+---+--------------------------+-------------+------------+-------------+
| # | NAME                     | POWER STATE | BRIGHTNESS | TEMPERATURE |
+---+--------------------------+-------------+------------+-------------+
| 0 | Elgato\ Key\ Light\ 861A | on          |         25 |         295 |
+---+--------------------------+-------------+------------+-------------+
[keylightctl(master)] $ 
```

