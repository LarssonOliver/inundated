export * from "./mapUtils";
export * from "./tagMapper";
export * from "./projectMapper";
export * from "./timespanMapper";
export * from "./settingsMapper";

export interface Mapper<Domain, Api> {
  fromApi(apiModel: Api): Domain;
  toApi(domainModel: Domain): Api;
}
