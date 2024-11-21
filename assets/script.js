// Generate the dynamic form using FRONTMATTER_DATA
function generateDynamicForm(frontmatter) {
  const form = document.getElementById("dynamic-form");
  form.innerHTML = ""; // Clear existing form content

  Object.entries(frontmatter).forEach(([key, value]) => {
    const label = document.createElement("label");
    label.textContent = key.charAt(0).toUpperCase() + key.slice(1);
    label.setAttribute("for", key);

    let input;

    if (typeof value === "boolean") {
      input = document.createElement("input");
      input.type = "checkbox";
      input.name = key;
      input.id = key;
      input.checked = value;
    } else if (typeof value === "string" && !isNaN(Date.parse(value))) {
      input = document.createElement("input");
      if (value.includes("T")) {
        input.type = "datetime-local";
      } else if (value.length === 10) {
        input.type = "date";
      } else {
        input.type = "time";
      }
      input.name = key;
      input.id = key;
      input.value = value;
    } else if (Array.isArray(value)) {
      const select = document.createElement("select");
      select.name = key;
      select.id = key;
      select.setAttribute("multiple", "multiple");
      value.forEach((item) => {
        const option = document.createElement("option");
        option.value = item;
        option.textContent = item;
        option.selected = true;
        select.appendChild(option);
      });
      input = select;
    } else {
      input = document.createElement("input");
      input.type = "text";
      input.name = key;
      input.id = key;
      input.value = value || "";
    }

    form.appendChild(label);
    form.appendChild(input);
    form.appendChild(document.createElement("br"));

    // Apply Select2 to multi-select fields
    if (input.tagName === "SELECT") {
      $(input).select2({
        tags: true,
        tokenSeparators: [",", " "],
      });
    }
  });
}

// Initialize EditorJS with BLOCKS_DATA
const editor = new EditorJS({
  holder: "editorjs",
  tools: {
    alert: Alert,
    code: CodeTool,
    delimiter: {
      class: Delimiter,
      config: {
        defaultLineWidth: 100,
        defaultStyle: "line",
        lineWidthOptions: [100],
        lineThicknessOptions: [2],
        styleOptions: ["line"],
      },
    },
    header: Header,
    image: ImageTool,
    list: EditorjsList,
    raw: RawTool,
    quote: Quote,
    table: Table,
  },
  data: {
    blocks: BLOCKS_DATA || [],
  },
});

// Generate the dynamic form
generateDynamicForm(FRONTMATTER_DATA);

// Handle form submission
document.getElementById("submit-button").addEventListener("click", async () => {
  const form = document.getElementById("dynamic-form");
  const formData = new FormData(form);
  const frontmatter = {};
  for (const [key, value] of formData.entries()) {
    const element = form.elements[key];
    if (element.type === "checkbox") {
      frontmatter[key] = element.checked;
    } else if (element.tagName === "SELECT") {
      frontmatter[key] = $(`#${key}`).val();
    } else {
      frontmatter[key] = value;
    }
  }

  try {
    const editorData = await editor.save();
    console.log("Frontmatter:", frontmatter);
    console.log("EditorJS Blocks:", editorData);
  } catch (error) {
    console.error("Error saving editor content:", error);
  }
});
