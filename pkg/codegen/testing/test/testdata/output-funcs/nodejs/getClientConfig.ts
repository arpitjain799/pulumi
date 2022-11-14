// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

/**
 * Failing example taken from azure-native. Original doc: Use this function to access the current configuration of the native Azure provider.
 */
export function getClientConfig(opts?: pulumi.InvokeOptions): Promise<GetClientConfigResult> {

    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("mypkg::getClientConfig", {
    }, opts);
}

/**
 * Configuration values returned by getClientConfig.
 */
export interface GetClientConfigResult {
    /**
     * Azure Client ID (Application Object ID).
     */
    readonly clientId: string;
    /**
     * Azure Object ID of the current user or service principal.
     */
    readonly objectId: string;
    /**
     * Azure Subscription ID
     */
    readonly subscriptionId: string;
    /**
     * Azure Tenant ID
     */
    readonly tenantId: string;
}
