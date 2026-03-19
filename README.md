# Модель Изинга на Go

Репозиторий содержит реализацию двумерной модели Изинга на решётке с методом Монте-Карло (алгоритм Метрополиса).

Проект реализован как полный вычислительный пайплайн:
- генерация входных данных,
- запуск симуляции,
- постобработка результатов.

---

## 📁 Структура проекта

```
.
├── cmd/
│   └── run/
│       └── main.go         # Точка входа: чтение input.csv и запись output.csv
├── ising/
│   └── ising.go            # Реализация модели Изинга
├── task-parser3.py         # Генерация input.csv из JSON
├── raw_data_converter.py   # Постобработка (C, X, Xafm)
├── run_simulation.bat      # Автоматический запуск всего пайплайна
├── params-sample2d.json    # Пример входных параметров
├── go.mod
└── README.md
```

---

## ⚙️ Полный пайплайн

```
JSON → task-parser3.py → input.csv → Go → output.csv → raw_data_converter.py → result.csv
```

---

## 📥 Формат входного файла `input.csv`

Разделитель — `;`

```
L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save
```

---

## 📤 Формат выходного файла `output.csv`

```
L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save;E;E2;Mtot;M2;Afm;Afm2
```

---

## 📊 Постобработка

Файл result.csv содержит:

```
T;C;X;Xafm
```

---

## ▶️ Запуск

Дважды кликнуть:

```
run_simulation.bat
```

Или вручную:

```
py task-parser3.py params-sample2d.json
go run ./cmd/run/main.go
py raw_data_converter.py output.csv result.csv
```
