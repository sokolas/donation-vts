# Installation

Go to https://github.com/sokolas/donation-vts/releases and download the latest release zip (`donation-vts(32-bit).zip` or `donation-vts(64-bit).zip`), unzip and run

# Usage

**Инструкция на русском (manual in russian)**: https://github.com/sokolas/donation-vts/blob/master/README.md

Edit `config.json` and set the following values:
*	`"logVtsMessages"`: `true`/`false` - log incoming messages from vtube studio (enable this to troubleshoot, otherwise you don't need it)
*	`"vtsAddr"`: `"ws://localhost:8001"` - your vtube studio address; see https://github.com/DenchiSoft/VTubeStudio/wiki/Plugins#how-to-use-plugins
*	`"vtsToken"`: `""` - leave it empty, the plugin sets it automatically. If you have authentication errors from vtube studio, set this to empty value manually (`""`)
*	`"customParam"`: `"DonationPluginParam"` - custom parameter name. It has to be unique and you need it to bind it to your model parameter later
*	`"paramDescription"`: `"donation alerts custom param"` - parameter description that is seen in vtube studio
*	`"stayTime"`: `20` (integer number) - each time you receive a donation, the parameter will be set for this duration in seconds before decaying to zero. If the value is negative the parameter will never decay.
*	`"decayTime"`: `5` (integer number) - the time for parameter value to decay to 0 (in seconds) after it's stay period is over. It can't be less than 1 second; if it's 0 or less it will be set to 1 instead
*	`"addParam"`: `true`/`false` - if set to `true`, donations will add their value to the current param value; if set to `false`, the value will be overwritten each time
*    `"multiplier"`: `1` (fractional number) - when you receive a donation, its value in your default currency will be multiplied by this `multiplier` and added/set to the custom parameter, trimming it to the range from 0 to 100
*	`"daToken"`: `""` - donationalerts token; leave it empty, it is set automatically. If you encounter donationalerts authentication errors, set this manually to `""`
*   `"daPort"`": `9696` - local port for donationalert auth; it's bound to the app id so don't change it unless you have your own app registered there
*   `"daAppId"`: `"10695"` - app id for donationalert auth. Don't change it unless you have your own app registered there


Start the app

Set up a mapping for plugin custom parameter to your model parameter in vtube studio: https://github.com/DenchiSoft/VTubeStudio/wiki/Plugins#what-are-custom-parameters

Authorize the app in donationalerts by opening the `Authorize Donationalerts` shortcut created in the app folder or by copy-pasting the link from the app logs into your browser (https://www.donationalerts.com/oauth/authorize?client_id=10695&redirect_uri=http%3A%2F%2Flocalhost%3A9696%2F&response_type=token&scope=oauth-donation-subscribe+oauth-user-show with default settings).

# Example
Let's assume you have `stayTime: 10`, `decayTime: 5`, `addParam: true` and `multiplier: 10` and your parameter controls your head size. By default it's set to 0 so your head is of its normal size. Then you receive a donation of 5 USD. Assuming your current currency is EUR and the conversion rate is 0.9 EUR per USD, you receive `5*0.9 = 4.5` EUR. This value is multiplied by `multiplier` (10) and the head size param is set to 45. After 10 seconds of staying at 45, it starts to shrink back to 0 over the duration of 5 seconds.
If you receive another donation during these 15 seconds from the first one, its value will be calculated and added (as configured by `addParam: true`) to the current value and the decay timer will be set back to 10 seconds.
