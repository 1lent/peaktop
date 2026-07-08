# peaktop

> Apple Silicon system monitor for the terminal — real-time CPU, GPU, memory, network, battery, and thermal metrics with zero configuration.
> <img width="1255" height="486" alt="image" src="https://github.com/user-attachments/assets/bf2e03a2-6ab4-420c-aac0-c2069539fdce" />


## Quick Start

```bash
go install github.com/1lent/peaktop@latest
peaktop
```

Or clone and build:

```bash
git clone https://github.com/1lent/peaktop.git
cd peaktop && make build && ./peaktop
```

## Usage

```bash
peaktop                 # 500ms update interval (default)
peaktop -i 200          # 200ms interval (faster)
peaktop -i 1000         # 1 second interval (slower)
sudo peaktop            # required for temperature, fan, and power metrics
```

## Keybinds

| Key | Action |
|---|---|
| `1`–`5` | Switch tabs |
| `t` | Cycle theme (Dark → Light → Dracula) |
| `+` / `-` | Increase / decrease tick rate |
| `j` / `k` or `↓` / `↑` | Scroll process list |
| `h` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit (with save/discard log prompt) |

## Tabs

**Overview** — CPU/GPU/ANE gauges with sparklines, per-core heatmap blocks, CPU/GPU frequency, VRAM usage, memory pressure bar, network per-interface throughput, thermal state, battery status, FPS counter.

**Processes** — Top 50 processes by CPU% with PID, name, CPU%, and MEM%.

**Thermal** — Thermal pressure level, CPU/GPU temperatures (sudo), fan speeds with activity bars and sparkline (sudo, M-series only).

**Network** — Per-interface RX/TX throughput updated in real-time.

**Battery** — Charge percentage bar, health, cycle count, time remaining, design capacity. Hidden on desktop systems without batteries.

## Features

- **CPU** — Per-core usage with P/E cluster averages, core heatmap blocks (red/yellow/green/grey), per-core labels, frequency, 60-sample sparkline history
- **GPU** — Usage gauge, active frequency, VRAM used/total, sparkline history
- **ANE** — Apple Neural Engine utilization (chip-dependent, shows "unavailable" when not exposed)
- **Memory** — Wired/active/compressed bar with breakdown, swap usage, pressure percentage
- **Network** — Total throughput plus per-interface RX/TX rates
- **Disk** — Read/write bytes per second, IOPS
- **Thermal** — Pressure level with percentage, CPU/GPU die temperatures, fan RPM bars with history sparkline
- **Battery** — Charge level, charging status, health (with effective mAh when degraded), cycle count, time remaining
- **Power** — Package/CPU/GPU/ANE/DRAM wattage breakdown (sudo required)
- **Alerts** — Configurable thresholds for thermal, battery, memory, GPU with cooldown periods
- **Themes** — Dark, Light, and Dracula — toggle at runtime or set in config
- **CSV Logging** — All metrics saved to `~/.peaktop/logs/YYYY-MM-DD.csv`, prompt to save or discard on quit

## CSV Logs

Every session writes to `~/.peaktop/logs/`. Each row contains:

`timestamp, cpu%, gpu%, ane%, power_w, cpu_w, gpu_w, ane_w, mem_used, mem_pressure%, swap_used, thermal_pressure, cpu_temp_c, gpu_temp_c, fan_rpm, battery%, charging, network_rx_bps, network_tx_bps`

On quit, you can press `y` to save or `n` to discard the session log.

## Config

Create `~/.peaktop/config.json`:

```json
{"theme": "dracula"}
```

Valid themes: `dark`, `light`, `dracula`. You can also cycle themes with the `t` key at runtime.

## Requirements

| Requirement | Details |
|---|---|
| OS | macOS 13+ |
| Architecture | Apple Silicon (arm64) |
| Go | 1.21+ (build from source only) |
| `sudo` | Only needed for temperature, fan, and power metrics |

## Platform Notes

**M-series Macs** (M1/M2/M3/M4) — Full feature set. Run with `sudo` for temperatures, fan speeds, and power metrics.

**A-series Neo devices** (A14–A18 Pro) — CPU, GPU, memory, network, disk, battery, and alerts work without sudo. Thermal pressure (`kern.thermalpressure`) is unavailable and shows "N/A". ANE may show "unavailable" if performance counters aren't exposed. These devices use passive cooling (no fans). Process visibility is limited compared to macOS.

## License

MIT
