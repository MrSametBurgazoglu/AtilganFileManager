# Atilgan File Manager

Atilgan is a lightweight and fast file manager for Linux, built with Go and the GTK4 toolkit. It provides a simple and intuitive interface for navigating and managing your files.

https://github.com/user-attachments/assets/6ec71452-fd24-41bd-aeb2-4b1af8bf5939

## Features

*   **File and Directory Listing:** Browse your files and directories in a list view.
*   **File Preview:** Preview various file types, including images, text files, documents, and videos.
*   **Path Navigation:** Easily navigate through your file system with a clickable path bar.
*   **Sidebar with Special Paths:** Quick access to your home directory, documents, downloads, and other special folders.
*   **Search:** Search for files and directories within the current directory.
*   **File Operations:** Perform common file operations like rename, copy, cut, and paste.
*   **Shortcuts:** A rich set of keyboard shortcuts for efficient navigation and management.
*   **Tags:** Organize your files with tags for easy categorization and search.

## Prerequisites

### Development Libraries

Before building Atilgan, you need to install the development libraries for GTK4, libadwaita, and gtksourceview.

**On Debian/Ubuntu:**
```bash
sudo apt-get install libgtk-4-dev libadwaita-1-dev gtksourceview5-dev
```

**On Fedora:**
```bash
sudo dnf install gtk4-devel libadwaita-devel gtksourceview5-devel
```

### Runtime Dependencies for Previews

For the file previewer to work correctly with all file types, you need to install the following packages:

*   **PDF Previews:** `pdftoppm` (usually part of the `poppler-utils` package)
    ```bash
    sudo apt-get install poppler-utils # Debian/Ubuntu
    sudo dnf install poppler-utils     # Fedora
    ```
*   **Document Previews:** `unoconv`
    ```bash
    sudo apt-get install unoconv # Debian/Ubuntu
    sudo dnf install unoconv     # Fedora
    ```
*   **Media Previews:** `ubuntu-restricted-extras` (for Ubuntu-based distributions)
    ```bash
    sudo apt-get install ubuntu-restricted-extras
    ```
    For other distributions, you may need to install GStreamer plugins for various media formats.

## Installation and Running

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/MrSametBurgazoglu/AtilganFileManager.git
    cd atilgan
    ```

2.  **Build the application:**
    ```bash
    go build
    ```

3.  **Run the application:**
    ```bash
    ./atilgan
    ```

## Shortcuts

| Shortcut      | Action                                       |
|---------------|----------------------------------------------|
| `Ctrl + R`    | Rename the selected file or directory.       |
| `Ctrl + F`    | Toggle the search bar.                       |
| `Ctrl + C`    | Copy the selected file or directory.         |
| `Ctrl + X`    | Cut the selected file or directory.          |
| `Ctrl + V`    | Paste the copied/cut file or directory.      |
| `Ctrl + H`    | Show the shortcuts help popup.               |
| `Escape`      | Clear the copied/cut files.                  |
| `Shift + [A-Z]` | Select the next file starting with the letter. |
| `Left Arrow`  | Go to the parent directory.                  |
| `Right Arrow` | Go into the selected directory.              |
