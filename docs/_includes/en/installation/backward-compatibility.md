werf follows a versioning strategy called [Semantic Versioning](https://semver.org). It means that major releases (1.0, 2.0) can break backward compatibility. In the case of werf, an update to the next major release _may_ require to do a full re-deploy of applications or to perform other non-scriptable actions.

Minor releases (1.1, 1.2, etc.) may introduce new global features, but have to do so without significant backward compatibility breaks with a major branch (1.x).
In the case of werf, this means that an update to the next minor release goes smoothly most of the time. However, it _may_ require running a provided upgrade script.

Patch releases (1.1.0, 1.1.1, 1.1.2) may introduce new features, but must do so without breaking backward compatibility within the minor branch (1.1.x).
In the case of werf, this means that an update to the next patch release should be smooth and can be done automatically.

- We do **not guarantee** backward compatibility between:
  - `alpha` releases;
  - `beta` releases;
  - `ea` releases.
- We **guarantee** backward compatibility between:
  - `stable` releases within the minor branch (1.1.x);
  - `rock-solid` releases within the minor branch (1.1.x).