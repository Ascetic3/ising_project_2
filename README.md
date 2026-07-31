# Модель Изинга на Go

Репозиторий содержит реализацию двумерной модели Изинга на решётке с методом Монте-Карло (алгоритм Метрополиса).

Проект реализован как полный вычислительный пайплайн:
- генерация входных данных,
- запуск симуляции,
- постобработка результатов.

---

## 📁 Структура демонстрационного запуска

```text
configs/demo-input.csv       # Проверенный небольшой входной CSV
cmd/run/main.go              # CLI и запись изолированных результатов
internal/csvio/input.go      # Parser входного CSV
ising/ising.go               # Модель и алгоритм Metropolis
tools/run_demo.bat           # Основной безопасный запуск v0.1
tools/run_simulation.bat     # Устаревшая безопасная заглушка
demo-output/                 # Игнорируемые Git результаты отдельных запусков
```

Пайплайн v0.1 не требует Python:

```text
configs/demo-input.csv → Go CLI → demo-output/<new-run>/
```

---

## 📥 Формат входного файла `input.csv`

Разделитель — `;`. Вход содержит ровно 13 полей в следующем порядке:

```text
L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save
```

Пример корректной строки:

```text
12;1;1;1;1;1;1;2;0;0.5;100;200;1
```

`save=1` продолжает эволюцию текущей решётки, а `save=0` сначала
восстанавливает ферромагнитную конфигурацию.

## 📤 Формат выходного файла `output.csv`

`output.csv` намеренно не содержит header для совместимости с существующей
постобработкой. Каждая строка содержит 13 входных полей и 6 измеренных величин:

```text
L;J1;J2;J3;J4;J5;J6;copies;h;T;aSteps;mSteps;save;E;E2;Mtot;M2;Afm;Afm2
```

`E`, `E2`, `Mtot`, `M2`, `Afm`, `Afm2` являются усреднёнными величинами,
возвращаемыми симулятором.

## 📊 `result.csv` и `diagnostics.csv`

В демонстрационном CLI `result.csv` создаётся непосредственно Go-программой и
не содержит header. Порядок колонок:

```text
T;E_per_spin;M_per_spin;AFM_per_spin;C;kappa;af_kappa
```

`diagnostics.csv` содержит header и одну строку на расчётную точку:

```text
point;T;point_seed
```

Python-скрипты в `scripts/` не участвуют в воспроизводимом demo v0.1.

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
