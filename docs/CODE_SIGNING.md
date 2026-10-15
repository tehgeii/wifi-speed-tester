# Code signing (removing the "Windows protected your PC" warning)

`WiFiSpeedTester.exe` is not code-signed yet. Windows SmartScreen therefore
warns on first run, and users have to click **More info → Run anyway**.
Signing proves the file comes from this project and has not been changed.

Signing cannot be done from the repository alone. It needs a certificate that
belongs to the project owner. This page lists the realistic options. Check
each provider's current terms before applying, because prices and eligibility
change.

## Options

| Option | Cost | Notes |
|---|---|---|
| **SignPath Foundation** (open-source program) | Free for qualifying open-source projects | Requires an OSI-approved open-source license, so it does **not** apply to this project's current license (see below). |
| **Microsoft Trusted/Artifact Signing** (Azure) | Monthly subscription | Cloud signing service with a GitHub Action. Identity validation is required, and availability for individual developers depends on the country. |
| **Commercial OV certificate** (Sectigo, DigiCert, …) | Yearly fee | The key must live on a hardware token or cloud HSM, so signing in CI needs the vendor's cloud-signing option. |

Even with a valid signature, SmartScreen builds "reputation" from downloads
over time. The warning goes away gradually, not always on day one.

## License and signing

Since version 1.3 the project uses its own license (`LICENSE`): the code is
public, but copying, modifying and selling need permission. That is not an
OSI-approved open-source license, so free programs for open-source projects
(such as SignPath Foundation) are not available. Microsoft Trusted/Artifact
Signing or a commercial certificate remain the options.

## How it would plug into the release workflow

Signing slots in between building the `.exe` and zipping it. With any
provider, the release job would:

1. build `WiFiSpeedTester.exe` (as `scripts/build.sh` does now);
2. send the `.exe` to the signing service (for example the SignPath or Azure
   GitHub Action) using secrets stored in **Settings → Secrets and
   variables → Actions**;
3. zip the **signed** `.exe` and publish the release.

Until a certificate exists, nothing here changes how releases are built.
