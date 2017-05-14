# Godown

Checks to see which sites in a provided list are down. An email will be sent that contains a list of the unresponsive sites along with associated error values.

## Install

1. Simplydownload the github repo and execute

```
> make && make install
```

## Edit `config.json` file

Create your config file by copying over the example file

```
> cp config.json.example ~/.config/godown/config.json
```

and editing it to suit your needs:

```
> vim config.json # insert gmail credentials, for example
> cat ~/.config/godown/config.json
```

