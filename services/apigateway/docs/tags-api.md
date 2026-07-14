# Теги и поиск по схожести

Теги доступны для родословных деревьев, кастомных деревьев и публичных людей. Клиент может выбирать только значения из общего справочника; пользовательские теги не создаются.

## Справочник

`GET /api/tags`

Ответ содержит `code`, русское `name` и `description`. Стабильный `code` используется во всех остальных запросах.

Доступные коды: `historical`, `contemporary`, `fictional`, `my_family`, `genealogy`, `dynasty`, `royalty`, `nobility`, `politics`, `military`, `science`, `education`, `academic`, `mentorship`, `medicine`, `technology`, `business`, `companies`, `organizations`, `religion`, `mythology`, `folklore`, `literature`, `books`, `films`, `tv_series`, `animation`, `anime_manga`, `comics`, `video_games`, `theatre`, `music`, `visual_arts`, `sports`, `local_history`, `cultural_heritage`, `migration`, `professional`.

## Назначение тегов

Во всех `PUT`-запросах тело одинаковое:

```json
{
  "tag_codes": ["historical", "royalty", "politics"]
}
```

Переданный набор полностью заменяет старый. Пустой массив снимает все теги. Дубликаты и пробелы нормализуются. Неизвестный код возвращает `400`. Изменять набор может только владелец.

| Объект | Получить | Заменить |
|---|---|---|
| Родословное дерево | `GET /api/familytree/{tree_id}/tags` | `PUT /api/familytree/{tree_id}/tags` |
| Кастомное дерево | `GET /api/custom-trees/{tree_id}/tags` | `PUT /api/custom-trees/{tree_id}/tags` |
| Публичный человек | `GET /api/public-persons/{public_person_id}/tags` | `PUT /api/public-persons/{public_person_id}/tags` |

Полные ответы дерева или публичного человека также содержат массив `tags`.

## Поиск

- Родословные: `GET /api/familytree/public/search?q=Романовы&tags=historical,royalty&limit=20`
- Кастомные деревья: `GET /api/custom-trees/public/search?q=МГУ&tags=academic&tags=mentorship&limit=20`
- Публичные люди: `GET /api/public-persons/search?q=Николай&tags=historical,royalty&limit=20`

Поддерживаются повторяемые `tags=...`, одиночный `tag=...` и несколько кодов через запятую. Текст `q` необязателен, когда выбран хотя бы один тег. Максимальный `limit` — 100.

Фильтр тегов работает как `OR`: объект должен иметь хотя бы один выбранный тег. Результаты сортируются по коэффициенту Жаккара:

```text
similarity = |выбранные ∩ теги объекта| / |выбранные ∪ теги объекта|
```

Поэтому полный или наиболее близкий набор идет раньше частичного. При равном `similarity_score` учитываются точное совпадение текста, совпадение начала названия и дата обновления. Без тегов сохраняется обычный текстовый поиск, а `similarity_score` равен `0`.
