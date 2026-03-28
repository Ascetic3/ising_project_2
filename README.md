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
├── internal/
│   └── csvio/
│       └── input.go        # Парсинг и валидация строк input.csv
├── scripts/
│   ├── make_input_csv.py   # Генерация input.csv из JSON
│   ├── make_result_csv.py  # Постобработка (C, X, Xafm)
│   └── graph_tool.py       # Внешний скрипт построения графиков (без изменений)
├── configs/
│   └── params-sample2d.json # Пример входных параметров
├── data/
│   ├── input/
│   │   └── input.csv       # Генерируется скриптом
│   └── output/
│       ├── output.csv      # Генерируется Go-симуляцией
│       ├── result.csv      # Генерируется постобработкой
│       └── plots/          # PNG-графики (run_simulation.bat делает cd сюда перед graph_tool)
├── tools/
│   └── run_simulation.bat  # Автоматический запуск всего пайплайна
├── go.mod
└── README.md
```

---

## ⚙️ Полный пайплайн

```
JSON → scripts/make_input_csv.py → data/input/input.csv → Go → data/output/output.csv → scripts/make_result_csv.py → data/output/result.csv
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
T;E;M;afm;C;X;Xafm
```

Скрипт `graph_tool.py` сохраняет PNG в текущую рабочую папку. Батник перед графиками делает `cd` в `data/output/plots`, поэтому файлы оказываются прямо в `data/output/plots/`. При ручном запуске из корня репозитория PNG появятся в корне.

---

## ▶️ Запуск

Дважды кликнуть:

```
tools/run_simulation.bat
```

Или вручную:

```
py scripts/make_input_csv.py configs/params-sample2d.json
go run ./cmd/run/main.go
py scripts/make_result_csv.py data/output/output.csv data/output/result.csv
py scripts/graph_tool.py data/output/result.csv
```

`tools/run_simulation.bat` автоматически проверяет наличие `numpy` и `matplotlib` перед шагом графиков.
Если библиотек нет, bat-файл пытается установить их автоматически.
Если установка не удалась, пайплайн завершается без графиков (с предупреждением).

При необходимости можно установить вручную:

```
py -m pip install matplotlib numpy
```
