export interface FrontmatterField {
  name: string;
  stringValue: string;
  boolValue: boolean;
  numberValue: number;
  dateTimeValue: string;
  stringSliceValue: string[];
  type: string;
}

export const DefaultFrontmatterFieldValues: FrontmatterField = {
  name: "",
  stringValue: "",
  boolValue: false,
  numberValue: 0,
  dateTimeValue: "0001-01-01T00:00:00Z",
  stringSliceValue: [],
  type: "",
};
