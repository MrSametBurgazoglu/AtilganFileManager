Name:           atilgan
Version:        0.2.0
Release:        1%{?dist}
Summary:        A modern, keyboard-centric file manager written in Go and GTK4

License:        GPLv3
URL:            https://github.com/MrSametBurgazoglu/AtilganFileManager
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  gtk4-devel
BuildRequires:  libadwaita-devel
BuildRequires:  libappstream-glib
BuildRequires:  desktop-file-utils
Requires:       gtk4
Requires:       libadwaita

%description
Atilgan File Manager is a fast, keyboard-driven file manager for Linux desktops.
It features a dual-pane layout, rich previews, and deep integration with GTK4 and Adwaita.

%prep
%autosetup

%build
go build -o atilgan .

%install
install -D -m 0755 atilgan %{buildroot}%{_bindir}/atilgan
install -D -m 0644 io.github.mrsametburgazoglu.AtilganFileManager.desktop %{buildroot}%{_datadir}/applications/io.github.mrsametburgazoglu.AtilganFileManager.desktop
install -D -m 0644 atilgan_icon.svg %{buildroot}%{_datadir}/icons/hicolor/scalable/apps/io.github.mrsametburgazoglu.AtilganFileManager.svg
install -D -m 0644 io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml %{buildroot}%{_datadir}/metainfo/io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml

%check
desktop-file-validate %{buildroot}%{_datadir}/applications/io.github.mrsametburgazoglu.AtilganFileManager.desktop
appstream-util validate-relax --nonet %{buildroot}%{_datadir}/metainfo/io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml

%files
%{_bindir}/atilgan
%{_datadir}/applications/io.github.mrsametburgazoglu.AtilganFileManager.desktop
%{_datadir}/icons/hicolor/scalable/apps/io.github.mrsametburgazoglu.AtilganFileManager.svg
%{_datadir}/metainfo/io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml

%changelog
* Sat May 09 2026 MrSametBurgazoglu <samet@example.com> - 0.2.0-1
- Initial RPM release
- Removed password storage for better security
- Removed Flatpak support
