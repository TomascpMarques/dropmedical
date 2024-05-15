# Dropmedical Backend

## Docker container instructions

No folder base do projeto, corremos o primeiro comando:

```bash
docker build -t dropmedical .
```

Este comando cria a imagem do contentor que contêm o serviço de api do dropmedical backend.

De seguida, corremos o comando a baixo na mesma pasta:

```bash
docker compose up
```

Este comando inicia dois serviços na mesma rede interna:

- DB -
  Uma instancia postgres com o porto 5432 mapeado para o porto 5431

- Api -
  A backend API do dropmedical, escrita em go, expões o porto 80. Logo, podemos aceder à api através do link: `http://localhost/api`
