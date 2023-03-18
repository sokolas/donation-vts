# Download
Latest version:
* https://github.com/sokolas/donation-vts/releases/download/v1.0.6/donation-vts.32-bit.zip 32-bit version
* https://github.com/sokolas/donation-vts/releases/download/v1.0.6/donation-vts.64-bit.zip 64-bit version

# Usage

**Русская версия**: https://github.com/sokolas/donation-vts/blob/master/README.md

First of all, enable plugins in vtube studio: [how-to](https://github.com/DenchiSoft/VTubeStudio/wiki/Plugins#how-to-use-plugins)
The plugin doesn't have an installer. Just unzip everything into a separate folder and run `donations-vts.exe`. If everything is fine, the console window will appear with the program logs, and after a few seconds the browser UI will open. Vtube studio will prompt you to allow the plugin in a few seconds; enable it, you have to do it once.
Once the plugin is authorized, Vtube Studio status will be `READY`, and the new parameter will appear in Vtube Studio (`DonationPluginParam` by default).
Next you have to set up a mapping for this custom parameter to a model parameter. More details on this step [here](https://github.com/DenchiSoft/VTubeStudio/wiki/Plugins#what-are-custom-parameters)

After that you need to authorize the plugin in donationalerts. **it is only possible when the app is running** with any of these methods:
* Click the `GRANT` link in the UI
* Open the `Authorize Donationalerts` shortcut that there is in the program folder
* Follow [this link](https://www.donationalerts.com/oauth/authorize?client_id=10695&redirect_uri=http%3A%2F%2Flocalhost%3A9696%2F&response_type=token&scope=oauth-donation-subscribe+oauth-user-show) (works with default settings)

Any one of these options will open Donationalerts website with authorization prompt. Allow it. If everything is OK  you'll see the "Success!" page, you can close this tab now. You need to do this authorization on the first launch, usually you don't need to authorize the app again.
If everything is OK the donationalerts status will be `READY`.

To test the integrations follow [this link](https://www.donationalerts.com/dashboard/activity-feed/donations) and click "add donation"; enter the sum and check "alert donation in widget"; confirm. The program log will show you " Received donation from Donations/..." message, it means all is well.

The plugin is configured either from the browser UI or by editing `config.json`. If you edit the config file you need to restart the app to apply new settings. You don't have to restart Vtube studio.
If the browser UI can't apply settings for some reason, restart the app. You don't have to close and reopen the UI, it will refresh the setting automatically in a few seconds.

**Vtube Studio parameter setup**
*	`Parameter name` (`"customParam"` in config): `"DonationPluginParam"` - custom parameter name that will be affected by the donations. It should be mapped to a model parameter, see the link above.
*	`Description` (`"paramDescription"` in config): `"donation alerts custom param"` - custom parameter description, doesn't affect anything but the text in Vtube Studio. May be useful if you have a lot of plugins.
*	`Stay time` (`"stayTime"` in config): `20` (integer number) - every time you receive a donation the custom parameter is set to a new value for **this duration** in seconds before starting to decay to 0. If the value is `-1` it never decays.
*	`Decay time` (`"decayTime"` in config): `5` (integer number) - the time (in seconds) it takes for the parameter to reduce from it's value to 0. See the example below.
*   `Multiplier` (`"multiplier"` in config): `1` (integer/float number) - when a donation is received its amount in your default currency is multiplied by this number before adding/setting the custom parameter value. See the example below.
*	`Add param value` (`"addParam"` in config): `true`/`false` - if checked(`true`), the donations will **add** the value to the param; if unchecked(`false`), then they will **replace** the current value.

**Other**
* `Reset tokens` - reset the authorizations for Vtube Studio and Donationalerts, use it if you see any access errors in the logs.
* `Show UI on startup` (`"autoOpenUi"` in config): `true`/`false` - open the browser UI on application startup. **if the tab is already open a new tab won't be created**

* `Save` - apply the settings and write the config. If they didn't apply or work, try restarting the plugin app. You don't need to restart Vtube Studio.
* `Reload` - discard your changes and read the current config from the plugin app.

Additional settings (**only change them if you really need to and you know what you're doing**)
*	`Log messages` (`"logVtsMessages"` for Vtube Studio, `"logDaMessages"` for donations in config): `true`/`false` - log the network messages. Use this if you face any issues and need help thourbleshooting.
*	`Address` (`"vtsAddr"` in config): `"ws://localhost:8001"` - Vtube Studio address for plugins. [More info about plugins](https://github.com/DenchiSoft/VTubeStudio/wiki/Plugins#how-to-use-plugins)
*	`"vtsToken"`: `""` - Vtube Studio plugin access token. The plugin app receives and stores it automatically. Only change/reset it if you encounter any access errors in the logs, but usually you should just click Reset tokens. You'll have to allow the plugin in Vtube Studio again.
*	`"daToken"`: `""` - donationalerts access token. **don't give this token to anyone!** The plugin app receives and stores it automatically. If you have any access errors in the logs, try clicking the Reset tokens button. You'll have to authorize the app in donationalerts again; see the usage above.
*   `"daPort"`": `9696` - port to run the UI and authorize donationalerts. **it is bound to the Donationalerts app ID and won't work if you change it**. If you have this port unavailable for some reason (another app is running on it) you'll have to [register your own app](https://www.donationalerts.com/application/clients) and set its port/id in the config.
*   `"daAppId"`: `"10695"` - AppID for Donationalerts authorization. **only change it if you have problems with the default settings and you know what you're doing**

# Example
Let's assume you have `stayTime: 10`, `decayTime: 5`, `addParam: true` and `multiplier: 10` and your parameter controls your head size. By default it's set to 0 so your head is of its normal size. Then you receive a donation of 5 USD. Assuming your current currency is EUR and the conversion rate is 0.9 EUR per USD, you receive `5*0.9 = 4.5` EUR. This value is multiplied by `multiplier` (10) and the head size param is set to 45. After 10 seconds of staying at 45, it starts to shrink back to 0 over the duration of 5 seconds.
If you receive another donation during these 15 seconds from the first one, its value will be calculated and added (as configured by `addParam: true`) to the current value and the decay timer will be set back to 10 seconds.
