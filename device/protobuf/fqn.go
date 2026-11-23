/*
 * Copyright (c) 2026 Brandon Miller-Mumford. Licensed under the MIT License; see LICENSE.
 */

package protobuf

// FQNApplyBillingHold is the name under which the device domain registers
// ApplyBillingHold and the billing hold rule dials it: the bare namespace, the
// domain and the operation DeviceService declares in device.proto. The serving
// domain's declaration and the dialling endpoint's rpc.Call both read it from
// here, so the two sides compile against one spelling and a rename fails at
// compile time in both trees; it sits beside the generated code because this
// module is the one both products already resolve and the generator writes no
// such constant. It carries no capability: the capability the operation requires
// is composed from the control realm in the shared security library.
const FQNApplyBillingHold = "orbtr.device.ApplyBillingHold"
