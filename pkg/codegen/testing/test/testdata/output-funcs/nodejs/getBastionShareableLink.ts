// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as utilities from "./utilities";

/**
 * Response for all the Bastion Shareable Link endpoints.
 * API Version: 2020-11-01.
 */
export function getBastionShareableLink(args: GetBastionShareableLinkArgs, opts?: pulumi.InvokeOptions): Promise<GetBastionShareableLinkResult> {

    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("mypkg::getBastionShareableLink", {
        "bastionHostName": args.bastionHostName,
        "resourceGroupName": args.resourceGroupName,
        "vms": args.vms,
    }, opts);
}

export interface GetBastionShareableLinkArgs {
    /**
     * The name of the Bastion Host.
     */
    bastionHostName: string;
    /**
     * The name of the resource group.
     */
    resourceGroupName: string;
    /**
     * List of VM references.
     */
    vms?: inputs.BastionShareableLink[];
}

/**
 * Response for all the Bastion Shareable Link endpoints.
 */
export interface GetBastionShareableLinkResult {
    /**
     * The URL to get the next set of results.
     */
    readonly nextLink?: string;
}

export function getBastionShareableLinkOutput(args: GetBastionShareableLinkOutputArgs, opts?: pulumi.InvokeOptions): pulumi.Output<GetBastionShareableLinkResult> {
    return pulumi.output(args).apply(a => getBastionShareableLink(a, opts))
}

export interface GetBastionShareableLinkOutputArgs {
    /**
     * The name of the Bastion Host.
     */
    bastionHostName: pulumi.Input<string>;
    /**
     * The name of the resource group.
     */
    resourceGroupName: pulumi.Input<string>;
    /**
     * List of VM references.
     */
    vms?: pulumi.Input<pulumi.Input<inputs.BastionShareableLinkArgs>[]>;
}
