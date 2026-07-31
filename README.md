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
JSON → scripts/make_input_csv.py → data/input/input.csv → Go (Union Jack Ising) → data/output/output.csv → scripts/make_result_csv.py → data/output/result.csv
```

---

## 📥 Формат входного файла `input.csv`

Разделитель — `;`

```
L;J1;J2;J3;J4;J5;J6;K;copies;h;T;aSteps;mSteps;save
```

---

## 📤 Формат выходного файла `output.csv`

```
L;J1;J2;J3;J4;J5;J6;K;copies;h;T;aSteps;mSteps;save;E;E2;Mtot;M2;Afm;Afm2
```

---

## 📊 Постобработка

Файл result.csv содержит:

```
T;E;M;afm;C;X;Xafm
```

Скрипт `graph_tool.py` сохраняет PNG в `./graphs` относительно текущей рабочей папки.

- При запуске через `tools/run_simulation.bat` (там выполняется `cd data/output/plots`) графики будут в `data/output/plots/graphs/`.
- При ручном запуске из корня репозитория графики будут в `./graphs` в корне проекта.

---

## 💾 Логика параметра `save`

В `cmd/run/main.go` используется следующая логика:

- `Simulator` пересоздаётся только если:
  - `sim == nil`
  - изменился `L`
  - изменился `copies`
- если `L` и `copies` не изменились:
  - `save = 1` → используется конфигурация предыдущей температуры
  - `save = 0` → вызывается `sim.ResetFerromagnetic()`



---

## ▶️ Запуск

Основной рекомендуемый запуск версии v0.1:

```bat
tools\run_demo.bat
```

Он создаёт новую изолированную директорию внутри `demo-output/` и не
перезаписывает предыдущие результаты. Python для демонстрационного запуска не
требуется.

`tools\run_simulation.bat` оставлен только как безопасная legacy-заглушка. Он
не запускает расчёт и не изменяет файлы; при вызове выводит ссылку на
`tools\run_demo.bat`.

---

## Reproducible demo

Requirements: Windows 10/11 and Go 1.22 or newer. Python packages are not
required for this demo.

Run this command from the repository root:

```bat
tools\run_demo.bat
```

The calculation normally finishes in a few seconds. Every run creates a new
timestamped directory under `demo-output/` and never overwrites an earlier
run. The directory contains:

- `input.csv` - an exact copy of the demo input;
- `output.csv` - raw Metropolis observables;
- `result.csv` - observables normalized for plotting or inspection;
- `diagnostics.csv` - point numbers, temperatures, and derived point seeds;
- `run_metadata.json` - the seed, model parameters, CLI arguments, Go version,
  and run time;
- `images/lattice_*.png` - red/blue snapshots of the spin lattice.

`tools\run_demo.bat` uses the fixed seed `20260731`. Reusing the same input and
seed reproduces the physical CSV and PNG files exactly; each independent copy
still receives its own deterministic random-number stream.

This small configuration demonstrates the qualitative change from a more
ordered lattice to a less ordered lattice. It is not a precision estimate of a
critical temperature or a replacement for a converged scientific simulation.
The public demo currently covers only the standard Metropolis calculation, not
parallel tempering.
