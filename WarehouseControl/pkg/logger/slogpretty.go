package logger

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func SetupPrettySlog() *slog.Logger {
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

/* <!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>WarehouseControl</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap');

  body {
    font-family: 'Inter', sans-serif;
    margin: 0; padding: 0;
    background: #f4f6f8;
    color: #333;
    transition: background-color 0.3s ease;
  }
  header {
    background: #2c3e50;
    padding: 15px 30px;
    color: #fff;
    display: flex;
    align-items: center;
    gap: 15px;
    font-weight: 600;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15);
  }
  label {
    font-weight: 600;
  }
  select, button {
    border-radius: 6px;
    border: 1px solid #ccc;
    padding: 7px 15px;
    font-size: 14px;
    outline-offset: 2px;
    transition: box-shadow 0.3s ease;
  }
  select:focus, button:focus {
    box-shadow: 0 0 6px #2980b9;
    border-color: #2980b9;
  }
  button {
    cursor: pointer;
    background: #3498db;
    color: white;
    border: none;
    transition: background-color 0.3s ease, transform 0.2s ease;
  }
  button:hover {
    background: #2980b9;
    transform: scale(1.05);
  }
  button:active {
    transform: scale(0.95);
  }
  #app {
    max-width: 1100px;
    margin: 30px auto;
    padding: 20px 20px 50px;
    background: white;
    border-radius: 10px;
    box-shadow: 0 6px 20px rgba(0,0,0,0.1);
    display: none;
    animation: fadeIn 0.6s ease forwards;
  }
  @keyframes fadeIn {
    from {opacity: 0; transform: translateY(20px);}
    to {opacity: 1; transform: translateY(0);}
  }
  h2 {
    margin-top: 0;
    font-weight: 700;
    color: #2c3e50;
    border-bottom: 2px solid #3498db;
    padding-bottom: 5px;
    margin-bottom: 15px;
  }
  table {
    border-collapse: collapse;
    width: 100%;
    margin-top: 15px;
    transition: box-shadow 0.3s ease;
  }
  table:hover {
    box-shadow: 0 2px 12px rgba(52, 152, 219, 0.25);
  }
  th, td {
    padding: 14px 15px;
    border-bottom: 1px solid #ddd;
    text-align: left;
    font-size: 15px;
    user-select: none;
  }
  th {
    background: #ecf0f1;
    font-weight: 600;
  }
  tbody tr:hover {
    background: #f1f7fb;
    cursor: pointer;
    transition: background-color 0.3s ease;
  }
  .readonly {
    color: #7f8c8d;
  }
  #add-item-btn {
    margin-top: 10px;
    margin-bottom: 15px;
    transition: background-color 0.3s ease, transform 0.2s ease;
  }
  #add-item-btn:hover {
    background-color: #27ae60;
    transform: scale(1.05);
  }
  #item-form {
    max-width: 400px;
    margin: 20px auto 40px;
    background: #ecf0f1;
    padding: 25px;
    border-radius: 10px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.12);
    display: none;
    animation: slideDown 0.5s ease forwards;
  }
  @keyframes slideDown {
    from {opacity: 0; transform: translateY(-20px);}
    to {opacity: 1; transform: translateY(0);}
  }
  #item-form h3 {
    margin-top: 0;
    color: #34495e;
  }
  #item-form label {
    display: block;
    margin-bottom: 18px;
    font-weight: 600;
    color: #2c3e50;
  }
  #item-form input[type="text"], #item-form input[type="number"] {
    width: 100%;
    padding: 10px 14px;
    border-radius: 6px;
    border: 1px solid #bdc3c7;
    font-size: 15px;
    transition: border-color 0.3s ease;
  }
  #item-form input[type="text"]:focus, #item-form input[type="number"]:focus {
    border-color: #2980b9;
    outline: none;
    box-shadow: 0 0 6px rgba(41, 128, 185, 0.6);
  }
  #item-form button {
    margin-right: 12px;
  }
  .button-danger {
    background: #e74c3c;
  }
  .button-danger:hover {
    background: #c0392b;
  }
  .flex-center {
    display: flex;
    justify-content: center;
    gap: 10px;
  }
</style>
</head>
<body>

<header>
  <label for="role-select">Выберите роль пользователя:</label>
  <select id="role-select" aria-label="Выберите роль пользователя">
    <option value="viewer">Просмотрщик</option>
    <option value="manager">Менеджер</option>
    <option value="admin">Администратор</option>
  </select>
  <button id="login-btn" aria-label="Войти в систему">Войти</button>
</header>

<div id="app" role="main" tabindex="-1">

  <h2>Список товаров</h2>
  <button id="add-item-btn" aria-label="Добавить новый товар">Добавить товар</button>
  <table id="items-table" aria-label="Таблица списка товаров">
    <thead>
      <tr><th>ID</th><th>Название</th><th>Количество</th><th>Действия</th></tr>
    </thead>
    <tbody></tbody>
  </table>

  <h2>История изменений</h2>
  <table id="history-table" aria-label="Таблица истории изменений">
    <thead>
      <tr><th>Дата</th><th>Пользователь</th><th>Действие</th><th>Товар ID</th><th>Описание изменений</th></tr>
    </thead>
    <tbody></tbody>
  </table>

  <button id="export-csv-btn" aria-label="Экспортировать историю изменений в CSV">Экспорт истории в CSV</button>

</div>

<div id="item-form" role="dialog" aria-modal="true" aria-labelledby="form-title" tabindex="-1">
  <h3 id="form-title">Добавить товар</h3>
  <input type="hidden" id="item-id" />
  <label for="item-name">Название:</label>
  <input type="text" id="item-name" aria-required="true" />

  <label for="item-qty">Количество:</label>
  <input type="number" id="item-qty" aria-required="true" />

  <div class="flex-center">
    <button id="save-item-btn">Сохранить</button>
    <button id="cancel-item-btn" class="button-danger">Отмена</button>
  </div>
</div>

<script>
  let currentRole = null;
  let items = [];
  let history = [];
  let editingItem = null;

  const roleSelect = document.getElementById('role-select');
  const loginBtn = document.getElementById('login-btn');
  const appDiv = document.getElementById('app');
  const itemsTableBody = document.querySelector('#items-table tbody');
  const historyTableBody = document.querySelector('#history-table tbody');
  const addItemBtn = document.getElementById('add-item-btn');
  const exportCsvBtn = document.getElementById('export-csv-btn');
  const itemFormDiv = document.getElementById('item-form');
  const formTitle = document.getElementById('form-title');
  const itemIdInput = document.getElementById('item-id');
  const itemNameInput = document.getElementById('item-name');
  const itemQtyInput = document.getElementById('item-qty');
  const saveItemBtn = document.getElementById('save-item-btn');
  const cancelItemBtn = document.getElementById('cancel-item-btn');

  function renderItems() {
    itemsTableBody.innerHTML = '';
    items.forEach(item => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${item.id}</td>
        <td>${item.name}</td>
        <td>${item.qty}</td>
        <td>
          ${(currentRole === 'admin' || currentRole === 'manager') ? `<button class="action-btn" onclick="editItem(${item.id})">Редактировать</button>` : ''}
          ${currentRole === 'admin' ? `<button class="action-btn button-danger" onclick="deleteItem(${item.id})">Удалить</button>` : ''}
        </td>
      `;
      if (currentRole === 'viewer') {
        tr.classList.add('readonly');
      }
      itemsTableBody.appendChild(tr);
    });
  }

  function renderHistory() {
    historyTableBody.innerHTML = '';
    history.forEach(h => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${h.date}</td>
        <td>${h.user}</td>
        <td>${h.action}</td>
        <td>${h.itemId}</td>
        <td>${h.desc}</td>
      `;
      historyTableBody.appendChild(tr);
    });
  }

  function addHistory(action, itemId, desc) {
    const now = new Date();
    const date = now.toLocaleString();
    history.push({date, user: currentRole, action, itemId, desc});
    renderHistory();
  }

  loginBtn.onclick = () => {
    currentRole = roleSelect.value;
    appDiv.style.display = 'block';
    setTimeout(() => {
      appDiv.focus();
    }, 100);
    renderItems();
    renderHistory();
    addItemBtn.style.display = (currentRole === 'admin' || currentRole === 'manager') ? 'inline-block' : 'none';
  };

  addItemBtn.onclick = () => {
    editingItem = null;
    formTitle.innerText = 'Добавить товар';
    itemIdInput.value = '';
    itemNameInput.value = '';
    itemQtyInput.value = '';
    itemFormDiv.style.display = 'block';
    itemNameInput.focus();
  };

  saveItemBtn.onclick = () => {
    const id = editingItem ? editingItem.id : (items.length ? Math.max(...items.map(i=>i.id)) + 1 : 1);
    const name = itemNameInput.value.trim();
    const qty = parseInt(itemQtyInput.value);
    if (!name || isNaN(qty)) {
      alert('Заполните все поля корректно');
      return;
    }

    if (editingItem) {
      editingItem.name = name;
      editingItem.qty = qty;
      addHistory('Редактирование', editingItem.id, `Название: ${name}, Кол-во: ${qty}`);
    } else {
      const newItem = {id, name, qty};
      items.push(newItem);
      addHistory('Добавление', id, `Название: ${name}, Кол-во: ${qty}`);
    }
    itemFormDiv.style.display = 'none';
    renderItems();
  };

  cancelItemBtn.onclick = () => {
    itemFormDiv.style.display = 'none';
  };

  window.editItem = function(id) {
    const item = items.find(i => i.id === id);
    if (!item) return;
    editingItem = item;
    formTitle.innerText = 'Редактировать товар';
    itemIdInput.value = item.id;
    itemNameInput.value = item.name;
    itemQtyInput.value = item.qty;
    itemFormDiv.style.display = 'block';
    itemNameInput.focus();
  };

  window.deleteItem = function(id) {
    if (confirm('Удалить товар №' + id + '?')) {
      items = items.filter(i => i.id !== id);
      addHistory('Удаление', id, 'Товар удалён');
      renderItems();
    }
  };

  exportCsvBtn.onclick = () => {
    let csvContent = "data:text/csv;charset=utf-8,";
    csvContent += 'Дата,Пользователь,Действие,Товар ID,Описание\n';
    history.forEach(row => {
      csvContent += [row.date, row.user, row.action, row.itemId, row.desc].join(",") + "\n";
    });
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", "history.csv");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

</script>

</body>
</html>
*/
