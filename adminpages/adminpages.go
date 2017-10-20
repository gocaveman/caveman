// Administrative control panel page and tools.
package adminpages

// need a way to register things - not sure if that's in here or a "adminregistry" subdir

// hm, gonna package in admin pages right here i guess... that should be fun - in this case
// adminpagesreg should probably be the thing to reference this and include it in the default
// handlers???? needs definite think-through.  probalby need to set up other default registries
// and see how it all compares; be sure to follow the same rule though - registries are just
// defaults but everything is actually wired in main.go
//
// interesting thought: there might only be one admin page - just the listing, and the rest is
// just items that are registered and we link to them
