package render

import (
    "fmt"
    "math"
    glm "github.com/Jragonmiris/mathgl"
    collada "github.com/GlenKelley/go-collada"
)

type Index struct {
    Collada *collada.Collada
    Id map[collada.Id]interface{}
    Data map[collada.Id]interface{}
    Mesh map[collada.Id]*Mesh
    Transforms map[collada.Id]glm.Mat4d
    VisualScene *collada.VisualScene
// animations: Object
// cameras: Object
// controllers: Object
// effects: Object
// geometries: Object
// images: Object
// lights: Object
// materials: Object
// scene: VisualScene
// visualScenes: Object
}

type Mesh struct {
   VerticesId string
   Polylist []*Polylist
}

type Polylist struct {
    VertexData []float64
    NormalData []float64
    TriangleElements []int16
}

func NewIndex(c *collada.Collada) (*Index, error) {
    index := &Index{
        c,
        make(map[collada.Id]interface{}),
        make(map[collada.Id]interface{}),
        make(map[collada.Id]*Mesh),
        make(map[collada.Id]glm.Mat4d),
        nil,
    }
    index.init()
    return index, nil
}

func (index *Index) AddId(id collada.Id, obj interface{}) {
    prev, ok := index.Id[id]
    if ok && prev != obj {
        panic("object id already defined: " + id)
    } else {
        index.Id[id] = obj
    }
}

func (index *Index) init() {
    index.indexVisualScenes()
    index.indexGeometry()
    
    ivs := index.Collada.Scene.InstanceVisualScene
    if ivs != nil {
        id, ok := ivs.Url.Id()
        if ok {
            index.VisualScene = index.Id[id].(*collada.VisualScene)
        }
    }
}

func NodeTransform(node *collada.Node) glm.Mat4d {
    transform := glm.Ident4d()
    for _, translate := range node.Translate {
        v := translate.F()
        transform = transform.Mul4(glm.Translate3Dd(v[0], v[1], v[2]))
    }
    for _, rotation := range node.Rotate {
        v := rotation.F()
        transform = transform.Mul4(glm.HomogRotate3Dd(v[3]*math.Pi/180, glm.Vec3d{v[0],v[1],v[2]}))
    }
    for _, scale := range node.Scale {
        v := scale.F()
        transform = transform.Mul4(glm.Scale3Dd(v[0],v[1],v[2]))
    }
    return transform
}

func (index *Index) indexVisualScenes() {
	for _, lib := range index.Collada.LibraryVisualScenes {
		for _, vs := range lib.VisualScene {
			if len(vs.Id) != 0 {
                index.AddId(vs.Id, vs)
			}
			for _, node := range vs.Node {
            index.indexNode(node)
			}
		}
	}
}

func (index *Index) indexNode(node *collada.Node) {
	if len(node.Id) != 0 {
      index.AddId(node.Id, node)
      index.Transforms[node.Id] = NodeTransform(node)
	}
   for _, child := range node.Node {
      index.indexNode(child)
   }
}

func (index *Index) indexGeometry() {
	for _, lib := range index.Collada.LibraryGeometries {
		for _, g := range lib.Geometry {
			if len(g.Id) != 0 {
                index.AddId(g.Id, g)
			}
         index.Mesh[g.Id] = index.createMesh(g.Mesh)
		}
	}
}

func (index *Index) createMesh(m *collada.Mesh) *Mesh {
    for _, source := range m.Source {
        index.Data[source.Id] = source.FloatArray.F()
    }
    verticies := map[string]collada.Id{}
    for _, i := range m.Vertices.Input {
        verticies[i.Semantic], _ = i.Source.Id()
    }
    mesh := &Mesh{
       string(m.Vertices.Id),
       make([]*Polylist, len(m.Polylist)),
    }
    for k, pl := range m.Polylist {
        mpl := index.createMeshPolyList(pl, verticies)
        mesh.Polylist[k] = mpl
    }
    return mesh
}

type IndexPair struct {
    V int
    N int
}

func (index *Index) ReadOffsets(inputs []*collada.InputShared, verticies map[string]collada.Id) (int, []float64, int, []float64, int) {
    skip := len(inputs)
    v := 0
    n := 1
    var vs []float64 = nil
    var ns []float64 = nil
    for _, input := range inputs {
        if input.Semantic == "VERTEX" {
            vid := verticies["POSITION"]
            vs = index.Data[vid].([]float64)
            v = int(input.Offset)
        } else if input.Semantic == "NORMAL" {
            nid, _ := input.Source.Id()
            ns = index.Data[nid].([]float64)
            n = int(input.Offset)
        }
    }
    return v, vs, n, ns, skip
}
func (index *Index) createMeshPolyList(pl *collada.Polylist, verticies map[string]collada.Id) *Polylist {
    used := map[IndexPair]int{}
    n := 0
    
    mpl := Polylist {
        make([]float64,0),
        make([]float64,0),
        make([]int16,0),
    }
    
    vo, vs, no, ns, stride := index.ReadOffsets(pl.Input, verticies)
    dimensions := 3
    
    //0 1 2 3
    //0 1 2 - 0 2 3
    p := pl.P.I()
    i := 0
    for _, v := range pl.VCount.I() {
        for j := 0; j < v; j++ {
            vertexIndex := p[i+vo]
            normalIndex := p[i+no]
            pair := IndexPair{vertexIndex, normalIndex}
            index, ok := used[pair]
            if !ok {
                index = n
                dv := vertexIndex*dimensions
                dn := normalIndex*dimensions
                for k := 0; k < 3; k++ {
                    mpl.VertexData = append(mpl.VertexData, vs[dv+k])
                    mpl.NormalData = append(mpl.NormalData, ns[dn+k])
                }
                used[pair] = index
                n++
            }
            switch v {
            case 3:
                mpl.TriangleElements = append(mpl.TriangleElements, int16(index))
             case 4:
                if j == 3 {
                   //split quad into triangle
                   n := len(mpl.TriangleElements)
                   p0 := mpl.TriangleElements[n-3]
                   p1 := mpl.TriangleElements[n-1]
                   mpl.TriangleElements = append(mpl.TriangleElements, p0, p1, int16(index))
                } else {
                   mpl.TriangleElements = append(mpl.TriangleElements, int16(index))
                }
            default:
                fmt.Println("unsupported polygon size:", v)
            }
            i += stride
        }
    }
    
    return &mpl
}

