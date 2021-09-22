// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package example

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PetType struct {
	Name *string `pulumi:"name"`
}

// PetTypeInput is an input type that accepts PetTypeArgs and PetTypeOutput values.
// You can construct a concrete instance of `PetTypeInput` via:
//
//          PetTypeArgs{...}
type PetTypeInput interface {
	pulumi.Input

	ToPetTypeOutput() PetTypeOutput
	ToPetTypeOutputWithContext(context.Context) PetTypeOutput
}

type PetTypeArgs struct {
	Name pulumi.StringPtrInput `pulumi:"name"`
}

func (PetTypeArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*PetType)(nil)).Elem()
}

func (i PetTypeArgs) ToPetTypeOutput() PetTypeOutput {
	return i.ToPetTypeOutputWithContext(context.Background())
}

func (i PetTypeArgs) ToPetTypeOutputWithContext(ctx context.Context) PetTypeOutput {
	return pulumi.ToOutputWithContext(ctx, i).(PetTypeOutput)
}

// PetTypeArrayInput is an input type that accepts PetTypeArray and PetTypeArrayOutput values.
// You can construct a concrete instance of `PetTypeArrayInput` via:
//
//          PetTypeArray{ PetTypeArgs{...} }
type PetTypeArrayInput interface {
	pulumi.Input

	ToPetTypeArrayOutput() PetTypeArrayOutput
	ToPetTypeArrayOutputWithContext(context.Context) PetTypeArrayOutput
}

type PetTypeArray []PetTypeInput

func (PetTypeArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]PetType)(nil)).Elem()
}

func (i PetTypeArray) ToPetTypeArrayOutput() PetTypeArrayOutput {
	return i.ToPetTypeArrayOutputWithContext(context.Background())
}

func (i PetTypeArray) ToPetTypeArrayOutputWithContext(ctx context.Context) PetTypeArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(PetTypeArrayOutput)
}

type PetTypeOutput struct{ *pulumi.OutputState }

func (PetTypeOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*PetType)(nil)).Elem()
}

func (o PetTypeOutput) ToPetTypeOutput() PetTypeOutput {
	return o
}

func (o PetTypeOutput) ToPetTypeOutputWithContext(ctx context.Context) PetTypeOutput {
	return o
}

func (o PetTypeOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v PetType) *string { return v.Name }).(pulumi.StringPtrOutput)
}

type PetTypeArrayOutput struct{ *pulumi.OutputState }

func (PetTypeArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]PetType)(nil)).Elem()
}

func (o PetTypeArrayOutput) ToPetTypeArrayOutput() PetTypeArrayOutput {
	return o
}

func (o PetTypeArrayOutput) ToPetTypeArrayOutputWithContext(ctx context.Context) PetTypeArrayOutput {
	return o
}

func (o PetTypeArrayOutput) Index(i pulumi.IntInput) PetTypeOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) PetType {
		return vs[0].([]PetType)[vs[1].(int)]
	}).(PetTypeOutput)
}

func init() {
	pulumi.RegisterOutputType(PetTypeOutput{})
	pulumi.RegisterOutputType(PetTypeArrayOutput{})
}