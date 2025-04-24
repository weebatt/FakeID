[//]: # (ОЧИЩАЕМ ВеСЬ КЭШ)
docker-compose down
docker system prune -f --all
docker builder prune

[//]: # (ОЧИЩАЕМ STAGE В GIT)
git clean -fd