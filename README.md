# ipruler

```bash
# ip link set eth2 up;
# ip route add to 172.31.201.2/24 dev eth2;
ip route add to default via 172.31.201.1 table 102;
ip rule add from 172.31.201.10 table 102;
ip rule add from 172.31.201.11 table 102;
```
