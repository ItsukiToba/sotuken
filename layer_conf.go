package neoimage

import (
	"fmt"
	"sort"
	"strings"
	"crypto/md5"
	"io"
	"os/exec"
	"os"
	"bytes"
	"math"
	"time"
	"sync"
	"context"

	// "strconv"
)

type ImageNeo struct {
	name string
	id string  //元のイメージとの対応付け
	files []string  //ファイル群
	layer_new []string  //マウントするディレクトリ
	layer_id []string
	layerNum int
	size []int
}

var (
	Imgs []ImageNeo
	rwMutex sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc
	once sync.Once
	isLayer = false
	prepareFlag = true
)

// const HOME = "/mnt/c/programing_practice/reserch/layer_config/"
const HOME = "/go/src/github.com/docker/docker/neoimage/"
var count = 0


func MountLayer(id string, upper string) {
	var img *ImageNeo
	for i, ig := range Imgs {
		if ig.layer_id[len(ig.layer_id)-1] == id {
			img = &Imgs[i]
			break
		}
	}

	overlay := "-t overlay overlay -o lowerdir="
	if len(img.layer_new) > 1 {
		for i:=len(img.layer_new)-1; i>0; i-- {
			overlay = overlay + "/var/lib/docker/overlay2/neoimage/" + img.layer_new[i] + ":"
		}
	}
	// for i:=0; i<count; i++ {
	// 	overlay = overlay + "/var/lib/docker/overlay2/neoimage/" + strconv.Itoa(i+1) + ":"
	// }
	overlay = overlay + "/var/lib/docker/overlay2/neoimage/" + img.layer_new[0] + ","
	overlay = overlay + "upperdir=/var/lib/docker/overlay2/neoimage/upper-" + img.id + ",workdir=/var/lib/docker/overlay2/neoimage/work-" + img.id + " " + upper
	cmd := exec.Command(HOME+"mountLayer.sh", overlay, img.id, upper)
	// cmd := exec.Command(HOME+"mountLayerMesure.sh", overlay, img.id, upper, strconv.Itoa(count))
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("Error make layer", err)
		os.Exit(1)
	}
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Println("Error make layer", err)
	// 	fmt.Printf("Combined Output:\n%s", string(output))
	// 	os.Exit(1)
	// }
	// count = count + 1
	/*--------------------test [the number of layer]-------------------*/
	// return "/var/lib/docker/overlay2/neoimage/merge-" + img.id
}

func prepareLayer() {
	if prepareFlag {
		prepareFlag = false
		ctx, cancel = context.WithCancel(context.Background())
		if !isLayer {
			ids := make([]string, 0)
			for i:=0; i<len(Imgs); i++ {
				Imgs[i].getAllFilePath()
				ids = append(ids, Imgs[i].id)
			}
			// Imgs[0].splitLayer()
			powerSet := powerSet(ids)
			for _, ps := range powerSet {
				common := getImage(ps[0]).files
				for i, id := range ps[1:] {
					common = getImage(id).getIntersection(common, ps[i])
				}
				size := calcSize(common, ps[0])
				divideFlag := true
				for _, id := range ps {
					if !getImage(id).RequiresSplit(size) {
						divideFlag = false	
						break			
					}
				}
				if len(ps) == 1 {
					divideFlag = true
				}
				if divideFlag {
					layerId := GenerateID()
					makeLayer(layerId, ps[0], common)
					for _, id := range ps {
						img := getImage(id)
						img.removeFile(common)
						img.layer_new = append(img.layer_new, layerId)
						img.layerNum = img.layerNum + 1
						img.size = append(img.size, size)
					}
				}
			}
	
			for _, img := range Imgs {
				umountPath := "/var/lib/docker/overlay2/neoimage/merge-test-"+img.id
				cmd := exec.Command(HOME+"umountLayerTest.sh", umountPath, img.id)
				_, err := cmd.Output()
				if err != nil {
					fmt.Println("Error umount test layer", err)
					os.Exit(1)
				}
			}
			isLayer = true
		}
		PrintLayerSize()
		cancel()
	}
}

func PrintLayerSize() {
	for _, img := range Imgs {
		fmt.Println()
		fmt.Printf("----------------------------------%s----------------------------------\n", img.name)
		for i:=0; i<len(img.layer_new); i++ {
			fmt.Printf("%s : %d byte\n", img.layer_new[i], img.size[i])
		}
		fmt.Println()
	}
}

func WaitPrepare() {
	prepareLayer()
	select {
	case <-ctx.Done():
		prepareFlag = true
		return
	}
}

func getImage(id string) *ImageNeo {
	for i, img := range Imgs {
		if img.id == id {
			return &Imgs[i]
		}
	}
	return nil
}

func SetImageName(mountedID string, imgName string) {
	for i, img := range Imgs {
		if img.layer_id[len(img.layer_id)-1] == mountedID {
			rwMutex.Lock()
			Imgs[i].name = imgName
			rwMutex.Unlock()	
		}
	}
}

func MakeImage(baseLayer string) {
	var img ImageNeo
	img.id = GenerateID()
	img.layer_id = append(img.layer_id, baseLayer)
	img.layerNum = 0
	rwMutex.Lock()
	Imgs = append(Imgs, img)
	rwMutex.Unlock()
}

func AddImage(addLayer string, parent string) {
	for i, img := range Imgs {
		if img.layer_id[len(img.layer_id)-1] == parent {
			rwMutex.Lock()
			Imgs[i].layer_id = append(Imgs[i].layer_id, addLayer)
			rwMutex.Unlock()
			return
		}
	}
	for _, img := range Imgs {
		for j, id := range img.layer_id {
			if id == parent {
				var ig ImageNeo
				ig.id = GenerateID()
				ig.layer_id = append(ig.layer_id, img.layer_id[:j+1]...)
				ig.layer_id = append(ig.layer_id, addLayer)
				ig.layerNum = 0
				rwMutex.Lock()
				Imgs = append(Imgs, ig)
				rwMutex.Unlock()
				break;
			}
		}
	}
}

func RemoveImage() {
	if isLayer {
		isLayer = false
	
		Imgs = nil
		
		cmd := exec.Command(HOME+"removeLayer.sh")
		_, err := cmd.Output()
		if err != nil {
			fmt.Println("Error make layer", err)
			os.Exit(1)
		}
	}
}

func PrintLayer(img ImageNeo) {
	for _, s := range img.layer_id {
		fmt.Println(s)
	}
}

// 関数: べき集合を計算
func powerSet(id []string) [][]string {
	n := len(id)
	input := make([]int, 0, n)
	for i:=0; i<n; i++ {
		input = append(input, i)
	} 
	// べき集合のサイズは 2^n
	powerSetSize := 1 << n
	result := make([][]int, 0, powerSetSize)

	// 各ビットパターンを調べて部分集合を生成
	for i := 0; i < powerSetSize; i++ {
		subset := []int{}
		for j := 0; j < n; j++ {
			// ビットが立っている場合に要素を追加
			if i&(1<<j) != 0 {
				subset = append(subset, input[j])
			}
		}
		result = append(result, subset)
	}
	ret := make([][]string, len(result))
	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) > len(result[j]) // 長さが大きい順
	})
	for i:=0; i<len(result); i++ {
		for j:=0; j<len(result[i]); j++ {
			ret[i] = append(ret[i], id[result[i][j]])
		}
	}
	ret = ret[:len(ret)-1]
	return ret
}


func hashFile(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer file.Close()
	hash := md5.New()
	
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return hash.Sum(nil)
}

// イメージ間で共通するファイルの集合を出力
func (in *ImageNeo)getIntersection(commonFile []string, compID string) []string {
	elementMap := make(map[string]bool)
	common := make([]string, 0)
	/* ファイル群の取得 */
	for _, f := range in.files {
		elementMap[f] = true
	}
	for _, f := range commonFile {
		if elementMap[f] {  //パスが同じ
			path1 :=  "/var/lib/docker/overlay2/neoimage/merge-test-" + in.id + f
			path2 :=  "/var/lib/docker/overlay2/neoimage/merge-test-" + compID + f
			fileInfo1, err := os.Lstat(path1)
			fileInfo2, err := os.Lstat(path2)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		
			// シンボリックリンクの判定
			if fileInfo1.Mode()&os.ModeSymlink!=0 && fileInfo2.Mode()&os.ModeSymlink!=0{
				target1, err := os.Readlink(path1)
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
			
				target2, err := os.Readlink(path2)
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
			
				// リンク先を比較
				if target1 == target2 {
					common = append(common, f)
				}
			}
		
			// 通常のファイルの判定
			if fileInfo1.Mode().IsRegular() && fileInfo2.Mode().IsRegular(){
				hash1 := hashFile(path1)
				hash2 := hashFile(path2)
				if bytes.Equal(hash1,hash2) {  //ファイル内容が同じ
					common = append(common, f)
				}
			}

			if fileInfo1.IsDir() && fileInfo2.IsDir() {
				common = append(common, f)
			}


		}
	}
	return common
}

// イメージに含まれるファイルの集合をフィールドに入れる
func (in *ImageNeo)getAllFilePath() {
	overlay := "-t overlay overlay -o lowerdir="
	if len(in.layer_id) > 1 {
		for i:=len(in.layer_id)-1; i>0; i-- {
			overlay = overlay + "/var/lib/docker/overlay2/" + in.layer_id[i] + "/diff:"
		}
	}
	overlay = overlay + "/var/lib/docker/overlay2/" + in.layer_id[0] + "/diff,"
	overlay = overlay + "upperdir=/var/lib/docker/overlay2/neoimage/upper-test-" + in.id + ",workdir=/var/lib/docker/overlay2/neoimage/work-test-" + in.id + " /var/lib/docker/overlay2/neoimage/merge-test-" + in.id
	cmd := exec.Command(HOME+"mountLayerTest.sh", overlay, in.id)
	// output, err := cmd.Output()
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error mount test layer", err)
		fmt.Printf("Combined Output:\n%s", string(output))
		os.Exit(1)
	}
	in.files = in.getFilePath("/var/lib/docker/overlay2/neoimage/merge-test-" + in.id)
}

// レイヤーidのファイル・リンク群を出力
func (in *ImageNeo)getFilePath(path string) []string {
	cmd := "sudo find " + path
	result, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		fmt.Println("Error:", err)
	}
	path_split := strings.Split(string(result), "\n")
	ret := make([]string, 0)
	for _, s := range path_split {
		if (s != "") {
			ret = append(ret, s[64:])
		}
	}
	return ret
}

// 渡されたファイルの合計を出力
func calcSize(files []string, id string) int {
	ret := 0
	for _, f := range files {
		filePath := "/var/lib/docker/overlay2/neoimage/merge-test-" + id + f
		fileInfo, err := os.Lstat(filePath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			ret = ret + int(fileInfo.Size())
		}
	
		// 通常のファイルの判定
		if fileInfo.Mode().IsRegular() {
			fileInfo, err = os.Stat(filePath)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			fileSize := int(fileInfo.Size()) // バイト単位のファイルサイズ
			ret = ret + fileSize
		}
	}
	return ret
}

// レイヤーを分けるべきかを判定する関数
func (in *ImageNeo)RequiresSplit(size int) bool {
	alpha := 1.0
	beta := 1.0
	Threshold := 100000.0
	
	v := alpha*float64(size) + beta*(fileIOModel(float64(in.layerNum+1))-fileIOModel(float64(in.layerNum)))  //評価式
	
	if v >= Threshold {
		return true
	} else {
		return false
	}
}

func fileIOModel(n float64) float64 {
	return 0.9*math.Exp(-1*n)+0.1
}

func makeLayer(layerId string, baseId string, files []string) {
	file, err := os.Create(HOME+"path.dat")
	if err != nil {
		fmt.Println("Error open file", err)
		os.Exit(1)
	}
	defer file.Close()

	for _, f := range files {
		// row := []string{f.path, f.layer}
		// 各行のスライスの要素をスペースで区切って1行にする
		// line := strings.Join(row, ",")

		// ファイルに書き込む
		_, err := file.WriteString(f + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
	cmd := exec.Command(HOME+"makeLayer.sh", layerId, baseId)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error make layer", err)
		fmt.Println("Output:", string(output))
		os.Exit(1)
	}
}

func GenerateID() string {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	return id
}

func (in *ImageNeo)removeFile(common []string) {
	// 作業用マップを作成して、slice2の要素をキーとして登録
	elementMap := make(map[string]bool)
	for _, f := range common {
		elementMap[f] = true
	}

	// 新しいスライスを作成し、slice1からslice2に含まれない要素を追加
	var result []string
	for _, f := range in.files {
		if !elementMap[f] {
			result = append(result, f)
		} 
	}
	in.files = result
}