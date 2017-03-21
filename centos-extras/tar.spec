Name:          tar
Version:       1.26
Release:       1
Summary:       A GNU file archiving program
Group:         Applications/Archiving
Vendor:        GNU
Distribution:  Centos5
Packager:      Davide Madrisan <davide.madrisan@...>
URL:           http://www.gnu.org/software/tar/tar.html
Source:        ftp://ftp.gnu.org/gnu/tar/%{name}-%{version}.tar.bz2
License:       GPL
BuildRequires: gettext-devel
BuildRoot:     %{_tmppath}/%{name}-%{version}-root

%description
The GNU tar program saves many files together into one archive and can restore individual files (or all of the files) from the archive.
Tar can also be used to add supplemental files to an archive and to update or list files in the archive.
Tar includes multivolume support, automatic archive compression/decompression, the ability to perform remote archives and the ability to perform incremental and full backups.
If you want to use Tar for remote backups, you'll also need to install the rmt package.
You should install the tar package, because you'll find its compression and decompression utilities essential for working with files.

%prep
%setup -q

%build
export FORCE_UNSAFE_CONFIGURE=1
%configure --bindir=/bin --libexecdir=/sbin

make

[ "%buildroot" != / ] && rm -rf "%buildroot"
%makeinstall

mv %{buildroot}/usr/bin %{buildroot}/bin
mv %{buildroot}/usr/libexec %{buildroot}/sbin

%find_lang %{name}

%clean
[ "%buildroot" != / ] && rm -rf "%buildroot"

%post
ln -s /bin/tar /bin/gtar
exit 0

%files -f %{name}.lang
%defattr(-,root,root,-)
/bin/tar
/sbin/rmt
/usr/share/info/dir
/usr/share/info/tar.info
/usr/share/info/tar.info-1
/usr/share/info/tar.info-2

%doc AUTHORS ChangeLog COPYING NEWS README THANKS TODO

%changelog
* Sat Sep 26 2011 Florian Travers <florian.traverse@...> 1.26-1
- changes for Centos5

* Sat Mar 12 2011 Automatic Build System <autodist@...> 1.26-1mamba
- automatic update by autodist

* Mon Nov 08 2010 Automatic Build System <autodist@...> 1.25-1mamba
- automatic update to 1.25 by autodist

* Mon Oct 25 2010 Automatic Build System <autodist@...> 1.24-1mamba
- automatic update to 1.24 by autodist

* Fri Mar 12 2010 Davide Madrisan <davide.madrisan@...> 1.23-1mamba
- update to 1.23

* Thu Mar 05 2009 Silvan Calarco <silvan.calarco@...> 1.22-1mamba
- automatic update to 1.22 by autodist

* Sat Dec 27 2008 Silvan Calarco <silvan.calarco@...> 1.21-1mamba
- automatic update to 1.21 by autodist

* Tue Apr 15 2008 Aleph0 <aleph0@...> 1.20-1mamba
- update to 1.20
- removed patch against CVE-2007-4476: merged upstream

* Thu Nov 29 2007 Aleph0 <aleph0@...> 1.18-2mamba
- fix against tar stack crashing in safer_name_suffix (CVE-2007-4476)

* Mon Jul 02 2007 Aleph0 <aleph0@...> 1.18-1mamba
- update to 1.18

* Mon Jun 11 2007 Aleph0 <aleph0@...> 1.17-1mamba
- update to 1.17

* Wed Jan 17 2007 Davide Madrisan <davide.madrisan@...> 1.16.1-1qilnx
- update to version 1.16.1 by autospec
- security update: fixes CVE-2006-6097
- dropped patch against CVE-2006-0300: fixed upstream

* Tue May 16 2006 Davide Madrisan <davide.madrisan@...> 1.15.1-4qilnx
- rebuilt

* Wed Mar 01 2006 Davide Madrisan <davide.madrisan@...> 1.15.1-3qilnx
- security update for CVE-2006-0300

* Fri Apr 08 2005 Davide Madrisan <davide.madrisan@...> 1.15.1-2qilnx
- added scripts to install/remove info files
- added manpage for tar created by mandrake people

* Wed Jan 05 2005 Davide Madrisan <davide.madrisan@...> 1.15.1-1qilnx
- update to version 1.15.1 by autospec

* Tue Dec 21 2004 Davide Madrisan <davide.madrisan@...> 1.15-1qilnx
- update to version 1.15 by autospec

* Wed Sep 01 2004 Davide Madrisan <davide.madrisan@...> 1.14-3qilnx
- moved executables into the /bin directory
- added a symbolic link for gtar
- added standard documentation files

* Tue Jul 27 2004 Davide Madrisan <davide.madrisan@...> 1.14-2qilnx
- trivial fixes in the specfile needed by the QiLinux distromatic parser

* Tue May 11 2004 Davide Madrisan <davide.madrisan@...> 1.14-1qilnx
- new version rebuild

* Tue Apr 22 2003 Silvan Calarco <silvan.calarco@...> 1.13-2qilnx
- relocation of info dir under usr/share/info

* Tue Apr 09 2003 Luca Tinelli <luca.tinelli@...> 1.13-1qilnx
- first build
