# peaktop

Apple Silicon system monitor for the terminal.

## Install

```bash
go install github.com/brodie/peaktop@latest
```

Or clone and build:

```bash
git clone https://github.com/brodie/peaktop.git
cd peaktop
make build
```

## Usage

```bash
peaktop              # runs with 500ms update interval
peaktop -i 1000      # 1 second interval
sudo peaktop         # enables power metrics (powermetrics requires root)
```

## Keybinds

| Key | Action |
|-----|--------|
| `1`-`5` | Switch tabs (Overview / Processes / Thermal / Network / Battery) |
| `j` / `k` | Scroll process list down / up |
| `t` | Cycle theme (Dark → Light → Dracula) |
| `+` / `-` | Faster / slower tick rate |
| `h` | Toggle help |
| `q` / `Ctrl+C` | Quit |

## Tabs

**Overview** — CPU/GPU/ANE gauges with sparklines, per-core activity blocks, CPU/GPU frequency, VRAM, per-interface network, memory pressure, thermal state, battery, FPS counter  
**Processes** — Top 50 processes by CPU%, with PID, name, CPU%, and MEM%  
**Thermal** — Thermal state, temperature history, fan speeds (SMC required), CPU timeline  
**Network** — Per-interface RX/TX throughput in real-time  
**Battery** — Charge %, health, cycle count, time remaining (hidden on desktops)

## Features

- Real-time CPU per-core and cluster (P/E) usage with heatmap blocks
- GPU utilization, frequency, and VRAM via IOKit
- Apple Neural Engine (ANE) monitoring (chip-dependent)
- Thermal state and die temperatures (requires sudo for powermetrics)
- Per-process CPU and memory usage
- Configurable alert thresholds with cooldown
- CSV history export to `~/.peaktop/logs/` (active by default, saves on quit)
- 3 themes (Dark, Light, Dracula) via `~/.peaktop/config.json` or `t` key
- Power metrics with `sudo` (CPU/GPU/ANE/DRAM watts via powermetrics)

## Config

Create `~/.peaktop/config.json`:

```json
{"theme": "dracula"}
```

## Requirements

- macOS 13+
- Apple Silicon (arm64)
- Go 1.22+
