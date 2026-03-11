# procc0113

## Compile

```
go build -o provider provider.go
go build -o consumer consumer.go
```

## Runtime

### Provider

Sem PDF/A:
```
./provider &
```
<br/>

Com PDF/A:
```
./provider pdfa &
```

### Clean
```
./consumer &
```

Limpar o ambiente
```
killall provider
killall consumer`
```