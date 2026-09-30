export * from "./mapUtils";
export * from "./tagMapper";
export * from "./projectMapper";
export * from "./timespanMapper";
export * from "./settingsMapper";
export * from "./taskMapper";

/** Maps API models to domain models, for a model the client only reads. */
export interface FromApiMapper<Domain, Api> {
  fromApi(apiModel: Api): Domain;
}

export interface Mapper<Domain, Api> extends FromApiMapper<Domain, Api> {
  toApi(domainModel: Domain): Api;
}
