export * from "./mapUtils";
export * from "./tagMapper";

export interface Mapper<Domain, Api> {
  fromApi(apiModel: Api): Domain;
  toApi(domainModel: Domain): Api;
}
