# 🚀 Atilgan File Manager

Atilgan is a fast, lightweight, and modern file manager for Linux, meticulously crafted using the **Go** programming language and the **GTK4** toolkit with **Libadwaita**. It offers a clean, keyboard-centric interface that focuses on a seamless navigation experience without unnecessary bloat.

[![Go Report Card](https://goreportcard.com/badge/github.com/MrSametBurgazoglu/AtilganFileManager)](https://goreportcard.com/report/github.com/MrSametBurgazoglu/AtilganFileManager)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/MrSametBurgazoglu/AtilganFileManager)](https://github.com/MrSametBurgazoglu/AtilganFileManager/releases)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

![Atilgan Screenshot](screenshots/screenshot.png)

## ⚡ Quick Install

The fastest way to install Atilgan on **Ubuntu, Debian, Fedora** and their derivatives:

```bash
curl -fsSL https://raw.githubusercontent.com/MrSametBurgazoglu/AtilganFileManager/main/install.sh | sh
```
*The script automatically handles dependencies (GTK4, Libadwaita) and sets up your application menu.*

---

## ✨ Key Features

### 🔍 Advanced Previews
Atilgan isn't just for browsing; it's for seeing. Our built-in previewer supports:
*   **Code & Text:** Syntax highlighting for hundreds of languages with **internal search** (Ctrl+F).
*   **Images & Graphics:** High-quality previews with metadata display.
*   **Media:** Integrated video and audio playback.
*   **Documents:** PDF and Office document previews (requires `poppler-utils` / `unoconv`).
*   **Smart Directories:** View image folders as a **Gallery Grid** or a detailed list.

### ⌨️ Keyboard-Centric Workflow
Designed for speed, Atilgan allows you to manage your files without leaving the home row:
*   **Instant Search:** Toggle search with `Ctrl+F` and navigate results instantly.
*   **Power Navigation:** Use `Arrow Keys` for deep diving and `Shift + [Letter]` to jump to specific files.
*   **Quick Actions:** Open a terminal in your current path with a single click or menu action.

### 📂 File Management & Organization
*   **Dual View Modes:** Toggle between detailed lists and visual grids.
*   **Trash with Restoration:** Preview trashed items with their original path and date, and restore them with one click.
*   **Tagging System:** Organize files across your system with custom color-coded tags.
*   **Breadcrumb Pathbar:** Interactive pathbar for fast jumping between parent folders.
*   **Multi-Tab Workspace:** Open multiple directories in tabs to manage complex file operations.

### 🛠️ Built with Modern Tech
*   **Libadwaita:** Follows the latest GNOME design patterns for a native "Adaptive" look.
*   **Go & GTK4:** Blazing fast performance with modern memory safety.
*   **Customizable:** Toggle hidden files, adjust sorting (Name, Time, Size, Type), and more.

---

## 📥 Manual Installation

### Pre-built Packages
Download native packages from the [GitHub Releases](https://github.com/MrSametBurgazoglu/AtilganFileManager/releases) page.

| Platform | Format | Link |
| :--- | :--- | :--- |
| **Ubuntu / Debian / Mint** | `.deb` | [Download latest](https://github.com/MrSametBurgazoglu/AtilganFileManager/releases/latest) |
| **Fedora / RHEL** | `.rpm` | [Download latest](https://github.com/MrSametBurgazoglu/AtilganFileManager/releases/latest) |

### Building from Source
```bash
# Clone the repo
git clone https://github.com/MrSametBurgazoglu/AtilganFileManager.git
cd AtilganFileManager

# Install and build
make build
sudo make install
```

---

## ⌨️ Common Shortcuts

| Shortcut      | Action                                       |
|---------------|----------------------------------------------|
| `Ctrl + R`    | Rename selected item                         |
| `Ctrl + F`    | Search in current folder                     |
| `Ctrl + C/X/V`| Copy / Cut / Paste                           |
| `Space`       | Open Quick Preview                           |
| `Ctrl + T`    | New Tab                                      |
| `Ctrl + H`    | Show all shortcuts                           |
| `Left/Right`  | Back / Forward (Into Folder)                 |

---

## 🤝 Contributing

Contributions are welcome! Whether it's a bug report, a new feature idea, or a translation, feel free to open an issue or submit a pull request.

## 📄 License

Atilgan is released under the **GPL-3.0** License. See the `LICENSE` file for details.
