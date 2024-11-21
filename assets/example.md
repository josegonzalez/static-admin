# Example Document

This document demonstrates the different types of markdown blocks supported by the EditorJS and Goldmark.

---

## Alerts

> [!NOTE]
> This is a **primary** alert.  
> It contains some **important** information for the user.

> [!TIP]
> A helpful piece of advice spread over multiple lines to showcase the formatting.

> [!CAUTION]
> <div align='center'>
> This is a **center-aligned** alert with HTML content.  
> Line breaks and alignment are preserved.
> </div>

---

## Lists

- Unordered item 1
- Unordered item 2
  - Unordered item 2: Nested item 1
  - Unordered item 2: Nested item 2

1. Ordered item 1
2. Ordered item 2
    - Ordered item 2: Sub-item 1
    - Ordered item 2: Sub-item 2

- [x] Checklist item 1 (completed)  
- [ ] Checklist item 2 (incomplete)

---

## Tables

| Name      | Age | City        |
|-----------|-----|-------------|
| Alice     | 30  | New York    |
| Bob       | 25  | Los Angeles |

---

## Quotes

> "The only way to do great work is to love what you do."  
> -- <caption>Steve Jobs</caption>

---

## Code Blocks

### Go Example

```go
func main() {
    fmt.Println("Hello, world!")
}
```

### Python Example

```python
def hello():
    print("Hello, world!")
```

---

## Images

![A beautiful sunset](https://images.unsplash.com/photo-1639056610940-d7e9b0af3a99?fm=jpg&q=60&w=3000&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8M3x8YmVhdXRpZnVsJTIwc3Vuc2V0fGVufDB8fDB8fHww)

<figure>
    <img src="https://hips.hearstapps.com/hmg-prod/images/sunrise-with-moai-royalty-free-image-1595416668.jpg" alt="A stunning sunrise" class="stretched" />
    <figcaption>A stunning sunrise</figcaption>
</figure>

---

## Paragraphs

This is a simple paragraph with **bold**, *italic*, and `code` formatting.

---

## Thematic Break

---

## Headings

### Heading Level 3

#### Heading Level 4

---
