%define debug_package %{nil}
%define pkgname {{cpkg_name}}
%define version {{cpkg_version}}
%define bindir {{cpkg_bindir}}
%define release {{cpkg_release}}
%define dist {{cpkg_dist}}
%define binary {{cpkg_binary}}
%define tarball {{cpkg_tarball}}
%define datadir {{cpkg_datadir}}
%define shortname {{cpkg_shortname}}

Name: %{pkgname}
Version: %{version}
Release: %{release}.%{dist}
Summary: The Choria Prometheus File Exporter
License: Apache-2.0
URL: https://choria.io
Group: System Tools
Packager: R.I.Pienaar <rip@devco.net>
Source0: %{tarball}
BuildRoot: %{_tmppath}/%{pkgname}-%{version}-%{release}-root-%(%{__id_u} -n)

%description
Exports metrics found on files in disk, also includes a friendly utility cronjobs
and other similar tools can use to write counters and gauges.

%prep
%setup -q

%build

%install
rm -rf %{buildroot}
%{__install} -d -m0755  %{buildroot}/etc/sysconfig
%{__install} -d -m0755  %{buildroot}/usr/lib/systemd/system
%{__install} -d -m0755  %{buildroot}/etc/logrotate.d
%{__install} -d -m0755  %{buildroot}%{bindir}
%{__install} -d -m0755  %{buildroot}/var/log
%{__install} -d -m1777  %{buildroot}%{datadir}
%{__install} -m0644 dist/%{pkgname}.service %{buildroot}/usr/lib/systemd/system/%{pkgname}.service
%{__install} -m0644 dist/sysconfig %{buildroot}/etc/sysconfig/%{pkgname}
%{__install} -m0644 dist/prometheus-file-exporter-logrotate %{buildroot}/etc/logrotate.d/%{pkgname}
%{__install} -m0755 %{binary} %{buildroot}%{bindir}/%{shortname}
touch %{buildroot}/var/log/%{pkgname}.log

%clean
rm -rf %{buildroot}

%post
if [ $1 -eq 1 ] ; then
  systemctl --no-reload preset %{pkgname} >/dev/null 2>&1 || :
fi

/bin/systemctl --system daemon-reload >/dev/null 2>&1 || :

if [ $1 -ge 1 ]; then
  /bin/systemctl try-restart %{pkgname} >/dev/null 2>&1 || :;
fi

%preun
if [ $1 -eq 0 ] ; then
  systemctl --no-reload disable --now %{pkgname} >/dev/null 2>&1 || :
fi

%files
%{bindir}/%{shortname}
/etc/logrotate.d/%{pkgname}
/usr/lib/systemd/system/%{pkgname}.service
%attr(640, nobody, nobody)/var/log/%{pkgname}.log
%attr(1777, root, root)%{datadir}
%config(noreplace) /etc/sysconfig/%{pkgname}

%changelog
* Tue Dec 26 2017 R.I.Pienaar <rip@devco.net>
- Initial Release
