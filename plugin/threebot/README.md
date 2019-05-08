## Installation
in plugin.cfg add `threebot:github.com/threefoldtech/threebot_coredns`

## Configuration

in Corefile

```
. {
    threebot ZONE {
        explorer EXPLORER_URL
    }
}

```

e.g

```
. {
    threebot grid.tf. {
        explorer "https://explorer.testnet.threefoldtoken.com"
    }
}
```

that will enable threebot against testnet explorer



