# Confluence CLI - acon

Workflow Patterns:

```
acon page create -t "Page Title" -f content.md -s SPACE
acon page create -t "Child Page" -f content.md -s SPACE --parent PAGE_ID
echo "# Heading\n\nContent here" | acon page create -t "Title" -s SPACE
echo "# Heading\n\nContent here" | acon page create -t "Title" -s SPACE -f -
cat document.md | acon page create -t "Title" -s SPACE
URL=$(acon page create -t "Title" -f content.md -s SPACE)
ID=$(acon page create -t "Title" -f content.md -s SPACE --json | jq -r '.id')

acon page update PAGE_ID -f updated.md
acon page update PAGE_ID -f content.md -m "Fixed typos"
acon page update PAGE_ID -t "New Title" -f content.md
echo "# Heading\n\nContent here" | acon page update PAGE_ID -f -

acon space list
acon space list --json | jq '.[].key'

acon page list -s SPACE
acon page list -s SPACE --limit 100
acon page list -s SPACE --sort title
acon page list -s SPACE --sort modified --desc

acon page list --parent PAGE_ID
acon page list --parent PAGE_ID --sort title

acon search "error handling"
acon search "API documentation" -s SPACE
acon search --title "README"
acon search --title "Architecture" -s SPACE
acon search --label documentation
acon search --label api-reference -s SPACE
acon search "query" --label docs --creator me
acon search --cql "type=page AND space=SPACE AND label=important"
acon search --cql "creator=currentUser() AND lastModified > now('-7d')"

acon search "query" --limit 50
CURSOR=$(acon search "query" --json | jq -r '.nextCursor // empty')
acon search "query" --cursor "$CURSOR"

acon debug md < document.md
acon page view PAGE_ID --json | jq -r '.body.storage.value' | acon debug storage

for id in PAGE_ID1 PAGE_ID2 PAGE_ID3; do
  acon page view "$id" > "page-$id.md"
done

acon search --label outdated --json | jq -r '.results[].content.id' | while read id; do
  echo "Processing $id"
  acon page view "$id"
done

acon page move PAGE_ID --parent NEW_PARENT_ID
acon page delete PAGE_ID
```
