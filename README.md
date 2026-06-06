# trsd

## Buildar o projeto

```make build```

## Subir o sistema gerenciador

```make start-manager```

## Subir o provider

### Com suporte a PDF/A

```ENABLE_PDFA=1 make start-provider```

### Sem suporte a PDF/A

```make start-provider```

## Testes

### Self-healing

Teste a geração de PDF/A : ```curl http://localhost:8082/pdfa```

Observe os logs do core e execute:
```docker logs -f trsd-core-1``` e
```docker logs -f trsd-execute-1```


Remova o container do provider ou suba-o sem suporte a PDF/A e veja as alterações na saída dos logs:
```docker rm -f trsd-execute-1``` e
```make start-provider```


