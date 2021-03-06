package app

// These are usage (help) texts shown when the app is started without the required arguments.
// Note: Using an indentation of two spaces as it provides a nice "look" in the console.
var (
	InitUsage = `Write an empty template for a configuration file in TOML format.
  
  The empty configuration file is a good starting point, but you will need to fill in the details of the different
  sections before autobot will be able to perform any work.`
	VersionUsage = `Display the current build version of autobot.`
	ServeUsage   = `Start autobot as a web server (micro service).

  The web service offers these endpoints:
  - GET /                    responds with a service status
  - GET /vehiclestore/status responds with a status of the vehicle store
  - GET /lookup              performs a vehicle lookup. Query params: country, hash, regnr or vin
  - PATCH /vehicle           enables/disables a vehicle based on the given operation
  
  Example of a vehicle lookup:
  - GET /lookup?regnr=BK33877&country=dk
  
  While the server is running, a scheduler will periodically check for new vehicle data from its source(s).
  This happens according to the cron-style time expression given in the config file.`
	SyncUsage = `Synchronise manually with a specific data source.

  The parameter "-p" (or "--provider") should specify the name of a provider to sync with.
  The name must match the provider name from the config file; that would be "[Providers.NAME]".
  Example:
    if the config file contains "[Providers.TEST]", among others, and you want to run a synchronisation with TEST,
    just use "-p TEST".
  Please be patient as synchronisation may take a long time.`
	LookupUsage = `Perform a vehicle lookup based on registration or VIN.

  Formatting is currently limited to a human readable format.`
	ClearUsage = `Clear the vehicle store of all data.

  You need to run the sync command again before any vehicle data will be available.`
	StatusUsage = `Show status of the vehicle store.

  Shows some useful stats such as number of vehicles, time of last synchronisation etc.`
	DisableUsage = `Disable a vehicle.
	
  The disabled vehicle will only appear in a lookup if the option "--disabled" is used.
  Disabling vehicles does not affect synchronisation.`
	EnableUsage = `Enable a vehicle.
  
  The reenabled vehicle will reappear in vehicle lookups as normal.
  Enabling vehicles does not affect synchronisation.`
	QueryUsage = `Query for vehicles.
  
  Searches for vehicles using various criteria for text matching and sorting.
  For now, only an upper limit of the number of vehicles to return, is supported.
  It's the intention to support multiple output formats, but currently, just a fixed CSV format is supported.`
)
