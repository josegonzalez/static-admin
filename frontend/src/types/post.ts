import { Block } from "./block";
import { FrontmatterField } from "./frontmatter";

export interface Post {
  id: string;
  path: string;
  frontmatter: FrontmatterField[];
  blocks: Block[];
}
